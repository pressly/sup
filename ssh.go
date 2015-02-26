package main

import (
	"bytes"
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
	User       string
	Host       string
	Agent      net.Conn
	Conn       *ssh.Client
	Session    *ssh.Session // embed this....?
	StdinPipe  io.WriteCloser
	StdoutPipe io.Reader
	Env        map[string]string
}

type Command struct {
	Input  string
	Output io.Reader
}

type ErrConnect struct {
	User   string
	Host   string
	Reason string
}

func (e ErrConnect) Error() string {
	return fmt.Sprintf(`Connect("%v@%v"): %v`, e.User, e.Host, e.Reason)
}

func (c *SSHClient) parseHost(host string) error {
	c.Host = host

	// Remove extra "ssh://" schema
	if c.Host[:6] == "ssh://" {
		c.Host = c.Host[6:]
	}

	if at := strings.Index(c.Host, "@"); at != -1 {
		c.User = c.Host[:at]
		c.Host = c.Host[at+1:]
		log.Printf("User from: %v", c.Host)
		log.Printf("User to: %v @ %v", c.User, c.Host)
	}

	// Add default user, if not set
	if c.User == "" {
		c.User = os.Getenv("USER")
		log.Printf("User OS: %v", c.User)
	}

	if strings.Index(c.Host, "/") != -1 {
		return ErrConnect{c.User, c.Host, "unexpected slash in the host URL"}
	}

	// Add default port, if not set
	if strings.Index(c.Host, ":") == -1 {
		c.Host += ":22"
	}

	log.Printf("parsed: <%v>@<%v>", c.User, c.Host)

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
			log.Println("PublicKey:", agentSigners[0].PublicKey())
			signers = append(signers, agentSigners...)
		}
	}

	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signers...),
		},
	}

	conn, err := ssh.Dial("tcp", c.Host, config)
	if err != nil {
		return ErrConnect{c.User, c.Host, err.Error()}
	}

	sess, err := conn.NewSession()
	if err != nil {
		return ErrConnect{c.User, c.Host, err.Error()}
	}
	c.Session = sess

	// TODO: test env variables this way....
	// probably wont work... so we need a pty

	// modes := ssh.TerminalModes{
	// 	ssh.ECHO:          0,     // disable echoing
	// 	ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
	// 	ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	// }
	// // Request pseudo terminal
	// if err := sess.RequestPty("xterm", 80, 40, modes); err != nil {
	// 	log.Fatalf("request for pseudo terminal failed: %s", err)
	// }

	// sess.Setenv("HI", "SUP")

	var b bytes.Buffer
	sess.Stdout = &b
	sess.Stderr = &b
	// err = sess.Start("HI=123 ls -la ; echo $X ; echo sup ; HI=wooooo echo $HI")
	err = sess.Start("ls -la")
	if err != nil {
		return ErrConnect{c.User, c.Host, err.Error()}
	}

	log.Println(b.String())

	sess.Wait()

	return nil
}

func (s *SSHClient) Wait() {
	s.Session.Wait()
}

func (s *SSHClient) Read(b []byte) (int, error) {
	return s.StdoutPipe.Read(b)
}

func (s *SSHClient) Write(b []byte) (int, error) {
	return s.StdinPipe.Write(b)
}
