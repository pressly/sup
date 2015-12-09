package sup

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Client is a wrapper over the SSH connection/sessions.
type SSHClient struct {
	Conn         *ssh.Client
	Sess         *ssh.Session
	User         string
	Host         string
	RemoteStdin  io.WriteCloser
	RemoteStdout io.Reader
	RemoteStderr io.Reader
	ConnOpened   bool
	SessOpened   bool
	Running      bool
	Env          string //export FOO="bar"; export BAR="baz";
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
	c.Host = host

	// Remove extra "ssh://" schema
	if len(c.Host) > 6 && c.Host[:6] == "ssh://" {
		c.Host = c.Host[6:]
	}

	if at := strings.Index(c.Host, "@"); at != -1 {
		c.User = c.Host[:at]
		c.Host = c.Host[at+1:]
	}

	// Add default user, if not set
	if c.User == "" {
		u, err := user.Current()
		if err != nil {
			return err
		}
		c.User = u.Username
	}

	if strings.Index(c.Host, "/") != -1 {
		return ErrConnect{c.User, c.Host, "unexpected slash in the host URL"}
	}

	// Add default port, if not set
	if strings.Index(c.Host, ":") == -1 {
		c.Host += ":22"
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
	files := []string{
		os.Getenv("HOME") + "/.ssh/id_rsa",
		os.Getenv("HOME") + "/.ssh/id_dsa",
	}
	for _, file := range files {
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

	fmt.Println("connecting", host)

	if c.ConnOpened {
		return fmt.Errorf("Already connected")
	}

	initAuthMethodOnce.Do(initAuthMethod)

	err := c.parseHost(host)
	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
	}

	c.Conn, err = dialer("tcp", c.Host, config)
	if err != nil {
		return ErrConnect{c.User, c.Host, err.Error()}
	}

	c.ConnOpened = true

	return nil
}

// Run runs the task.Run command remotely on c.Host.
func (c *SSHClient) Run(task *Task) error {
	if c.Running {
		return fmt.Errorf("Session already running")
	}
	if c.SessOpened {
		return fmt.Errorf("Session already connected")
	}

	sess, err := c.Conn.NewSession()
	if err != nil {
		return err
	}

	c.RemoteStdin, err = sess.StdinPipe()
	if err != nil {
		return err
	}

	c.RemoteStdout, err = sess.StdoutPipe()
	if err != nil {
		return err
	}

	c.RemoteStderr, err = sess.StderrPipe()
	if err != nil {
		return err
	}

	c.Sess = sess
	c.SessOpened = true

	// Start the remote command.
	if err := c.Sess.Start(c.Env + "set -x;" + task.Run); err != nil {
		return ErrTask{task, err.Error()}
	}

	c.Running = true
	return nil
}

// Wait waits until the remote command finishes and exits.
// It closes the SSH session.
func (c *SSHClient) Wait() error {
	if !c.Running {
		return fmt.Errorf("Trying to wait on stopped session")
	}

	err := c.Sess.Wait()
	c.Sess.Close()
	c.Running = false
	c.SessOpened = false

	return err
}

// DialThrough will create a new connection from the ssh server sc is connected to. DialThrough is an SSHDialer.
func (sc *SSHClient) DialThrough(net, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	fmt.Println("Dialing", net, addr)
	conn, err := sc.Conn.Dial(net, addr)
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
	if c.SessOpened {
		c.Sess.Close()
		c.SessOpened = false
	}
	if !c.ConnOpened {
		return fmt.Errorf("Trying to close the already closed connection")
	}

	err := c.Conn.Close()
	c.ConnOpened = false
	c.Running = false

	return err
}

func (c *SSHClient) Prefix() string {
	return c.User + "@" + c.Host
}

func (c *SSHClient) Write(p []byte) (n int, err error) {
	return c.RemoteStdin.Write(p)
}

func (c *SSHClient) WriteClose() error {
	return c.RemoteStdin.Close()
}
