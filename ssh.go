package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// SSHClient is a wrapper over the SSH connection/sessions.
type SSHClient struct {
	User         string
	Host         string
	Agent        net.Conn
	Conn         *ssh.Client
	Sess         *ssh.Session
	RemoteStdin  io.WriteCloser
	RemoteStdout io.Reader
	RemoteStderr io.Reader
	//TODO: Use Session RequestPty, Shell() and Session.Env()
	//Env      map[string]string
	Env        string //export FOO="bar"; export BAR="baz";
	ConnOpened bool
	SessOpened bool
	Running    bool
	Prefix     string
}

// parseHost parses and normalizes <user>@<host:port> from a given string.
func (c *SSHClient) parseHost(host string) error {
	c.Host = host

	// Remove extra "ssh://" schema
	if c.Host[:6] == "ssh://" {
		c.Host = c.Host[6:]
	}

	if at := strings.Index(c.Host, "@"); at != -1 {
		c.User = c.Host[:at]
		c.Host = c.Host[at+1:]
	}

	// Add default user, if not set
	if c.User == "" {
		c.User = os.Getenv("USER")
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

// Connect creates SSH connection to a specified host.
// It expects the host of the form "[ssh://]host[:port]".
func (c *SSHClient) Connect(host string) error {
	if c.ConnOpened {
		return fmt.Errorf("Already connected")
	}

	if err := c.parseHost(host); err != nil {
		return err
	}

	// TODO: add the keys from ~/ssh/config ..
	// Look for IdentityFiles .. etc...

	var signers []ssh.Signer

	// If there's a running SSH Agent, use its Private keys
	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err == nil {
		agent := agent.NewClient(sock)
		agentSigners, err := agent.Signers()
		if err == nil && len(agentSigners) > 0 {
			signers = append(signers, agentSigners...)
		}
	}

	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signers...),
		},
	}

	c.Conn, err = ssh.Dial("tcp", c.Host, config)
	if err != nil {
		return ErrConnect{c.User, c.Host, err.Error()}
	}
	c.ConnOpened = true

	return nil
}

// reconnect creates new session for the SSH connection.
func (c *SSHClient) reconnect() error {
	if c.SessOpened {
		return fmt.Errorf("Session already connected")
	}

	sess, err := c.Conn.NewSession()
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

	c.RemoteStdin, err = sess.StdinPipe()
	if err != nil {
		return err
	}

	c.Sess = sess
	c.SessOpened = true
	return nil
}

// Run runs the cmd.Exec command remotely on cmd.Host.
func (c *SSHClient) Run(cmd Command) error {
	if c.Running {
		return fmt.Errorf("Session already running")
	}

	// Reconnect session.
	if err := c.reconnect(); err != nil {
		return ErrConnect{c.User, c.Host, err.Error()}
	}

	// Start the remote command.
	// Pass `export FOO="bar"; export BAR="baz";` in front of the command.
	if err := c.Sess.Start(c.Env + cmd.Exec); err != nil {
		return ErrCmd{cmd, err.Error()}
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

	return err
}
