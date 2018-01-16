package sup

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Client is a wrapper over the SSH connection/sessions.
type SSHClient struct {
	conn         *ssh.Client
	sess         *ssh.Session
	user         string
	host         string
	remoteStdin  io.WriteCloser
	remoteStdout io.Reader
	remoteStderr io.Reader
	connOpened   bool
	sessOpened   bool
	running      bool
	env          string //export FOO="bar"; export BAR="baz";
	color        string
}

type ErrConnect struct {
	User   string
	Host   string
	Reason string
}

func (e ErrConnect) Error() string {
	return fmt.Sprintf(`Connect("%v@%v"): %v`, e.User, e.Host, e.Reason)
}

// parseHost parses and normalizes <user>@<host:port> from a given string.
func (c *SSHClient) parseHost(host string) error {
	c.host = host

	// Remove extra "ssh://" schema
	if len(c.host) > 6 && c.host[:6] == "ssh://" {
		c.host = c.host[6:]
	}

	if at := strings.Index(c.host, "@"); at != -1 {
		c.user = c.host[:at]
		c.host = c.host[at+1:]
	}

	// Add default user, if not set
	if c.user == "" {
		u, err := user.Current()
		if err != nil {
			return err
		}
		c.user = u.Username
	}

	if strings.Index(c.host, "/") != -1 {
		return ErrConnect{c.user, c.host, "unexpected slash in the host URL"}
	}

	// Add default port, if not set
	if strings.Index(c.host, ":") == -1 {
		c.host += ":22"
	}

	return nil
}

var initAuthMethodOnce sync.Once
var authMethod ssh.AuthMethod

// initAuthMethod initiates SSH authentication method.
func initAuthMethod() {
	var signers []ssh.Signer

	// If there's a running SSH Agent, try to use its Private keys.
	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err == nil {
		agent := agent.NewClient(sock)
		signers, _ = agent.Signers()
	}

	// Try to read user's SSH private keys form the standard paths.
	files, _ := filepath.Glob(os.Getenv("HOME") + "/.ssh/id_*")
	for _, file := range files {
		if strings.HasSuffix(file, ".pub") {
			continue // Skip public keys.
		}
		data, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}
		signer, err := ssh.ParsePrivateKey(data)
		if err != nil {
			continue
		}
		signers = append(signers, signer)

	}
	authMethod = ssh.PublicKeys(signers...)
}

// SSHDialFunc can dial an ssh server and return a client
type SSHDialFunc func(net, addr string, config *ssh.ClientConfig) (*ssh.Client, error)

// Connect creates SSH connection to a specified host.
// It expects the host of the form "[ssh://]host[:port]".
func (c *SSHClient) Connect(host string) error {
	return c.ConnectWith(host, ssh.Dial)
}

// ConnectWith creates a SSH connection to a specified host. It will use dialer to establish the
// connection.
// TODO: Split Signers to its own method.
func (c *SSHClient) ConnectWith(host string, dialer SSHDialFunc) error {
	if c.connOpened {
		return fmt.Errorf("Already connected")
	}

	initAuthMethodOnce.Do(initAuthMethod)

	err := c.parseHost(host)
	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User: c.user,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
	}

	c.conn, err = dialer("tcp", c.host, config)
	if err != nil {
		return ErrConnect{c.user, c.host, err.Error()}
	}
	c.connOpened = true

	return nil
}

// Run runs the task.Run command remotely on c.host.
func (c *SSHClient) Run(task *Task) error {
	if c.running {
		return fmt.Errorf("Session already running")
	}
	if c.sessOpened {
		return fmt.Errorf("Session already connected")
	}

	sess, err := c.conn.NewSession()
	if err != nil {
		return err
	}

	c.remoteStdin, err = sess.StdinPipe()
	if err != nil {
		return err
	}

	c.remoteStdout, err = sess.StdoutPipe()
	if err != nil {
		return err
	}

	c.remoteStderr, err = sess.StderrPipe()
	if err != nil {
		return err
	}

	if task.TTY {
		// Set up terminal modes
		modes := ssh.TerminalModes{
			ssh.ECHO:          0,     // disable echoing
			ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
			ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		}
		// Request pseudo terminal
		if err := sess.RequestPty("xterm", 80, 40, modes); err != nil {
			return ErrTask{task, fmt.Sprintf("request for pseudo terminal failed: %s", err)}
		}
	}

	// Start the remote command.
	if err := sess.Start(c.env + task.Run); err != nil {
		return ErrTask{task, err.Error()}
	}

	c.sess = sess
	c.sessOpened = true
	c.running = true
	return nil
}

// Wait waits until the remote command finishes and exits.
// It closes the SSH session.
func (c *SSHClient) Wait() error {
	if !c.running {
		return fmt.Errorf("Trying to wait on stopped session")
	}

	err := c.sess.Wait()
	c.sess.Close()
	c.running = false
	c.sessOpened = false

	return err
}

// DialThrough will create a new connection from the ssh server sc is connected to. DialThrough is an SSHDialer.
func (sc *SSHClient) DialThrough(net, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	conn, err := sc.conn.Dial(net, addr)
	if err != nil {
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil

}

// Close closes the underlying SSH connection and session.
func (c *SSHClient) Close() error {
	if c.sessOpened {
		c.sess.Close()
		c.sessOpened = false
	}
	if !c.connOpened {
		return fmt.Errorf("Trying to close the already closed connection")
	}

	err := c.conn.Close()
	c.connOpened = false
	c.running = false

	return err
}

func (c *SSHClient) Stdin() io.WriteCloser {
	return c.remoteStdin
}

func (c *SSHClient) Stderr() io.Reader {
	return c.remoteStderr
}

func (c *SSHClient) Stdout() io.Reader {
	return c.remoteStdout
}

func (c *SSHClient) Prefix() (string, int) {
	host := c.user + "@" + c.host + " | "
	return c.color + host + ResetColor, len(host)
}

func (c *SSHClient) Write(p []byte) (n int, err error) {
	return c.remoteStdin.Write(p)
}

func (c *SSHClient) WriteClose() error {
	return c.remoteStdin.Close()
}

func (c *SSHClient) Signal(sig os.Signal) error {
	if !c.sessOpened {
		return fmt.Errorf("session is not open")
	}

	switch sig {
	case os.Interrupt:
		// TODO: Turns out that .Signal(ssh.SIGHUP) doesn't work for me.
		// Instead, sending \x03 to the remote session works for me,
		// which sounds like something that should be fixed/resolved
		// upstream in the golang.org/x/crypto/ssh pkg.
		// https://github.com/golang/go/issues/4115#issuecomment-66070418
		c.remoteStdin.Write([]byte("\x03"))
		return c.sess.Signal(ssh.SIGINT)
	default:
		return fmt.Errorf("%v not supported", sig)
	}
}
