package client

import (
	"fmt"
	"io"
	"os/exec"
)

// Client is a wrapper over the SSH connection/sessions.
type LocalhostClient struct {
	Cmd     *exec.Cmd
	Stdin   io.WriteCloser
	Stdout  io.Reader
	Stderr  io.Reader
	Running bool
	Env     string //export FOO="bar"; export BAR="baz";
}

// cmd := exec.Command("/bin/sh", mongoToCsvSH)

func (c *LocalhostClient) Connect(host string) error {
	return nil
}

func (c *LocalhostClient) Run(task Task) error {
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
	return "whateveruser" + "@localhost"
}
