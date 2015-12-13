package sup

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/pressly/prefixer"
	"golang.org/x/crypto/ssh"
)

const VERSION = "0.2.2"

type Stackup struct {
	conf *Supfile
}

func New(conf *Supfile) (*Stackup, error) {
	return &Stackup{
		conf: conf,
	}, nil
}

// Run runs set of commands on multiple hosts defined by network sequentially.
// TODO: This megamoth method needs a big refactor and should be split
//       to multiple smaller methods.
func (sup *Stackup) Run(network *Network, commands ...*Command) error {
	if len(commands) == 0 {
		return errors.New("no commands to be run")
	}

	// Process all ENVs into a string of form
	// `export FOO="bar"; export BAR="baz";`.
	env := ``
	for name, value := range sup.conf.Env {
		env += `export ` + name + `="` + value + `";`
	}
	for name, value := range network.Env {
		env += `export ` + name + `="` + value + `";`
	}

	var paddingLen int

	// Create clients for every host (either SSH or Localhost).
	var (
		clients []Client
		bastion *SSHClient
	)

	if network.Bastion != "" {
		bastion = &SSHClient{}
		if err := bastion.Connect(network.Bastion); err != nil {
			return err
		}
	}

	for _, host := range network.Hosts {
		var c Client

		if host == "localhost" { // LocalhostClient

			local := &LocalhostClient{
				env: env + `export SUP_HOST="` + host + `";`,
			}
			if err := local.Connect(host); err != nil {
				return err
			}

			c = local

		} else { // SSHClient

			remote := &SSHClient{
				env: env + `export SUP_HOST="` + host + `";`,
			}

			var err error
			if bastion != nil {
				err = remote.ConnectWith(host, bastion.DialThrough)
			} else {
				err = remote.Connect(host)
			}
			if err != nil {
				return err
			}
			defer remote.Close()

			c = remote
		}

		len := len(c.Prefix())
		if len > paddingLen {
			paddingLen = len
		}

		clients = append(clients, c)
	}

	// Run command or run multiple commands defined by target sequentially.
	for _, cmd := range commands {
		// Translate command into task(s).
		tasks, err := TasksFromConfigCommand(cmd, env)
		if err != nil {
			return fmt.Errorf("TasksFromConfigCommand(): %s", err)
		}

		// Run tasks sequentally.
		for _, task := range tasks {

			var taskClients chan Client
			if task.RunOnce {
				taskClients = make(chan Client, 1)
				// Range over one client - range over map for randomness.
				for _, client := range clients {
					taskClients <- client
					break
				}
				close(taskClients)
			} else {
				// Range over all clients.
				taskClients = make(chan Client, len(clients))
				for _, client := range clients {
					taskClients <- client
				}
				close(taskClients)
			}

			var writers []io.Writer
			var wg sync.WaitGroup
			i := 0

			// Run tasks on the provided clients.
			for c := range taskClients {
				padding := strings.Repeat(" ", paddingLen-(len(c.Prefix())))
				color := Colors[i%len(Colors)]
				i++
				prefix := color + padding + c.Prefix() + " | "

				err := c.Run(task)
				if err != nil {
					log.Fatalf("%sexit %v", prefix, err)
				}

				// Copy over tasks's STDOUT.
				wg.Add(1)
				go func(c Client) {
					defer wg.Done()
					_, err := io.Copy(os.Stdout, prefixer.New(c.Stdout(), prefix))
					if err != nil && err != io.EOF {
						// TODO: io.Copy() should not return io.EOF at all.
						// Upstream bug? Or prefixer.WriteTo() bug?
						log.Printf("%sSTDOUT: %v", c.Prefix(), err)
					}
				}(c)

				// Wait for each client to finish the command.
				wg.Add(1)
				go func(c Client) {
					defer wg.Done()
					if err := c.Wait(); err != nil {
						//TODO: Handle the SSH ExitError in ssh pkg
						e, ok := err.(*ssh.ExitError)
						if ok && e.ExitStatus() != 15 {
							// TODO: Prefix should be with color.
							// TODO: Store all the errors, and print them after Wait().
							fmt.Fprintf(os.Stderr, "%s | exit %v\n", c.Prefix(), e.ExitStatus())
							os.Exit(e.ExitStatus())
						}
						// TODO: Prefix should be with color.
						fmt.Fprintf(os.Stderr, "%s | %v\n", c.Prefix(), err)
						os.Exit(1)
					}
				}(c)

				// Copy over tasks's STDERR.
				wg.Add(1)
				go func(c Client) {
					defer wg.Done()
					_, err := io.Copy(os.Stderr, prefixer.New(c.Stderr(), prefix))
					if err != nil && err != io.EOF {
						log.Printf("%sSTDERR: %v", c.Prefix(), err)
					}
				}(c)

				writers = append(writers, c.Stdin())
			}

			// Copy over task's STDIN.
			if task.Input != nil {
				writer := io.MultiWriter(writers...)
				_, err := io.Copy(writer, task.Input)
				if err != nil {
					log.Printf("STDIN: %v", err)
				}
				//TODO: Use MultiWriteCloser (not in Stdlib), so we can writer.Close()?
				// 	    Move this at least to some defer function instead.
				for _, c := range clients {
					c.WriteClose()
				}
			}

			wg.Wait()
		}
	}

	return nil
}
