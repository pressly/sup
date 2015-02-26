package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SSHClient struct { // SSHSession ...?
	User         string
	Host         string
	Agent        net.Conn
	Conn         *ssh.Client
	Sess         *ssh.Session // embed this....?
	RemoteStdin  io.WriteCloser
	RemoteStdout io.Reader
	RemoteStderr io.Reader
	Env          map[string]string
	Opened       bool
	Started      bool
}

type ErrConnect struct {
	User   string
	Host   string
	Reason string
}

func (e ErrConnect) Error() string {
	return fmt.Sprintf(`Connect("%v@%v"): %v`, e.User, e.Host, e.Reason)
}

type ErrCmd struct {
	Cmd    Command
	Reason string
}

func (e ErrCmd) Error() string {
	return fmt.Sprintf(`Run("%v"): %v`, e.Cmd, e.Reason)
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

// Connect connects SSH client to a specified host.
// Expects host in the format "[ssh://]host[:port]", returns error otherwise.
func (c *SSHClient) Connect(host string) error {
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

	return nil
}

func (c *SSHClient) reconnect() error {

	//TODO: Check for old sessions, error out if not closed?

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
	c.Opened = true
	return nil
}

func (c *SSHClient) Run(cmd Command) error {
	// Reconnect session
	if err := c.reconnect(); err != nil {
		return ErrConnect{c.User, c.Host, err.Error()}
	}

	// // =========== TODO: RequestPTY?
	// modes := ssh.TerminalModes{
	// 	//ssh.ECHO:          1,     // disable echoing
	// 	ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
	// 	ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	// }
	// // Request pseudo terminal
	// if err := sess.RequestPty("xterm", 80, 40, modes); err != nil {
	// 	conn.Close()
	// 	return ErrConnect{host, fmt.Sprintf("request for pseudo terminal failed: %s", err)}
	// }

	// =========== TODO: ENV
	// err = sess.Start("HI=123 ls -la ; echo $X ; echo sup ; HI=wooooo echo $HI")
	//err = sess.Start("echo $FOO")

	// for name, value := range c.Env {
	// 	if err := sess.Setenv(name, value); err != nil {
	// 		return ErrConnect{host, fmt.Sprintf(`Setenv("%v", "%v"): %v`, name, value, err.Error())}
	// 	}
	// }

	if err := c.Sess.Start(cmd.Exec); err != nil {
		return ErrCmd{cmd, err.Error()}
	}
	c.Started = true

	// if err := sess.Shell(); err != nil {
	// 	return ErrConnect{host, err.Error()}
	// }

	return nil
}

func (c *SSHClient) Wait() {
	if !c.Started {
		log.Printf("Trying to run Wait() on stopped SSHClient")
		return
	}
	c.Sess.Wait()
}

func (c *SSHClient) Close() {
	if !c.Opened {
		log.Printf("Trying to run Close() on closed SSHClient")
		return
	}
	c.Conn.Close()
}

func (c *SSHClient) Read(b []byte) (int, error) {
	log.Printf("Chci cist.... %s", b)
	return c.RemoteStdout.Read(b)
}

func (c *SSHClient) Write(b []byte) (int, error) {
	log.Printf("Chci psat.... %s", b)
	return c.RemoteStdin.Write(b)
}
