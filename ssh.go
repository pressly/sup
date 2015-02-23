package main

import (
	"bytes"
	"io"
	"log"
	"net"
	"os"

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
}

type Command struct {
	Input  string
	Output io.Reader
}

func (s *SSHClient) Connect(host string) error {
	var signers []ssh.Signer

	// TODO: add the keys from ~/ssh/config ..
	// Look for IdentityFiles .. etc...

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

	s.User = "ubuntu"
	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signers...),
		},
	}

	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		log.Println("naw..")
		return err
	}
	// defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		log.Println("naw..2")
		return err
	}
	s.Session = sess

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
		return err
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
