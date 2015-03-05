package client

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Client is a wrapper over the SSH connection/sessions.
type LocalhostClient struct {
	Cmd     *exec.Cmd
	User    string
	Stdin   io.WriteCloser
	Stdout  io.Reader
	Stderr  io.Reader
	Running bool
	Env     string //export FOO="bar"; export BAR="baz";
}

// cmd := exec.Command("/bin/sh", mongoToCsvSH)

func (c *LocalhostClient) Connect(_ string) error {
	c.User = os.Getenv("USER")
	return nil
}

func (c *LocalhostClient) Run(task Task) error {
	var err error

	if c.Running {
		return fmt.Errorf("Command already running")
	}

	cmd := exec.Command("bash", "-xc", task.Run)
	c.Cmd = cmd

	c.Stdout, err = cmd.StdoutPipe()
	if err != nil {
		return err
	}

	c.Stderr, err = cmd.StderrPipe()
	if err != nil {
		return err
	}

	c.Stdin, err = cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := c.Cmd.Start(); err != nil {
		return ErrTask{task, err.Error()}
	}

	c.Running = true

	return nil
}

func (c *LocalhostClient) Wait() error {
	if !c.Running {
		return fmt.Errorf("Trying to wait on stopped command")
	}
	err := c.Cmd.Wait()
	c.Running = false
	return err
}

func (c *LocalhostClient) Close() error {
	return nil
}

func (c *LocalhostClient) Prefix() string {
	return c.User + "@localhost"
}
