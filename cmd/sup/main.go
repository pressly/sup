package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/pressly/stackup/config"
	"github.com/pressly/stackup/ssh"
	"github.com/pressly/stackup/terminal"

	"github.com/pressly/prefixer"

	gossh "golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

// usage prints help for an arg and exits.
func usage(conf *config.Config, arg int) {
	log.Println("Usage: sup <network> <target/command>\n")
	switch arg {
	case 1:
		// <network> missing, print available hosts.
		log.Println("Available networks (from Supfile):")
		for name, network := range conf.Networks {
			log.Printf("- %v\n", name)
			for _, host := range network.Hosts {
				log.Printf("   - %v\n", host)
			}
		}
	case 2:
		// <target/command> not found or missing,
		// print available targets/commands.
		log.Println("Available targets (from Supfile):")
		for target, _ := range conf.Targets {
			log.Printf("- %v\n", target)
		}
		log.Println("Available commands (from Supfile):")
		for cmd, _ := range conf.Commands {
			log.Printf("- %v\n", cmd)
		}
	}
	os.Exit(1)
}

// parseArgs parses os.Args for network and commands to be run.
func parseArgs(conf *config.Config) (config.Network, []config.Command) {
	var commands []config.Command

	// Check for the first argument first
	if len(os.Args) < 2 {
		usage(conf, len(os.Args))
	}
	// Does the <network> exist?
	network, ok := conf.Networks[os.Args[1]]
	if !ok {
		log.Printf("Unknown network \"%v\"\n\n", os.Args[1])
		usage(conf, 1)
	}

	// Does <network> have any hosts?
	if len(network.Hosts) == 0 {
		log.Printf("No hosts specified for network \"%v\"", os.Args[1])
		usage(conf, 1)
	}

	// Check for the second argument
	if len(os.Args) < 3 {
		usage(conf, len(os.Args))
	}
	// Does the <target/command> exist?
	target, isTarget := conf.Targets[os.Args[2]]
	if isTarget {
		// It's the target. Loop over its commands.
		for _, cmd := range target {
			// Does the target's command exist?
			command, isCommand := conf.Commands[cmd]
			if !isCommand {
				log.Printf("Unknown command \"%v\" (from target \"%v\": %v)\n\n", cmd, os.Args[2], target)
				usage(conf, 2)
			}
			command.Name = cmd
			commands = append(commands, command)
		}
	} else {
		// It's probably a command. Does it exist?
		command, isCommand := conf.Commands[os.Args[2]]
		if !isCommand {
			// Not a target, nor command.
			log.Printf("Unknown target/command \"%v\"\n\n", os.Args[2])
			usage(conf, 2)
		}
		command.Name = os.Args[2]
		commands = append(commands, command)
	}

	// Check for extra arguments
	if len(os.Args) != 3 {
		usage(conf, len(os.Args))
	}

	return network, commands
}

func main() {
	var (
		conf       config.Config
		paddingLen int
	)

	// Read configuration file.
	data, _ := ioutil.ReadFile("./Supfile")
	if err := yaml.Unmarshal(data, &conf); err != nil {
		log.Fatal(err)
	}

	// Parse network and commands to be run from os.Args.
	network, commands := parseArgs(&conf)

	// Process all ENVs into a string of form
	// `export FOO="bar"; export BAR="baz";`.
	env := ``
	for name, value := range conf.Env {
		env += `export ` + name + `="` + value + `";`
	}
	for name, value := range network.Env {
		env += `export ` + name + `="` + value + `";`
	}

	// Open SSH connection to all the hosts.
	clients := make([]*ssh.Client, len(network.Hosts))
	for i, host := range network.Hosts {
		c := &ssh.Client{
			Env: env,
		}
		if err := c.Connect(host); err != nil {
			log.Fatal(err)
		}
		defer c.Close()

		len := len(c.User) + 1 + len(c.Host)
		if len > paddingLen {
			paddingLen = len
		}

		clients[i] = c
	}

	// Run the command(s) remotely on all hosts in parallel.
	// Run multiple commands (from) sequentally.
	for _, cmd := range commands {
		tasks, err := ssh.TasksFromConfigCommand(cmd)
		if err != nil {
			log.Fatalf("TasksFromConfigCommand(): ", err)
		}

		for _, task := range tasks {
			log.Printf("Running task %v", task)

			// Run the command on all hosts in parallel.
			for i, c := range clients {
				padding := strings.Repeat(" ", paddingLen-(len(c.User)+1+len(c.Host)))
				color := terminal.Colors[i%len(terminal.Colors)]

				c.Prefix = color + padding + c.User + "@" + c.Host + " | "
				c.Run(task)

				go func(c *ssh.Client) {
					if _, err := io.Copy(os.Stdout, prefixer.New(c.RemoteStdout, c.Prefix)); err != nil {
						log.Printf("%sSTDOUT error: %v", c.Prefix, c.Host, err)
					}
				}(c)
				go func(c *ssh.Client) {
					if _, err := io.Copy(os.Stderr, prefixer.New(c.RemoteStderr, c.Prefix)); err != nil {
						log.Printf("%sSTDERR error: %v", c.Prefix, c.Host, err)
					}
				}(c)
			}

			// Wait for all hosts to finish.
			for _, c := range clients {
				if err := c.Wait(); err != nil {
					//TODO: Handle the SSH ExitError in ssh pkg
					e, ok := err.(*gossh.ExitError)
					if !ok {
						log.Fatalf("%sexpected *ExitError but got %T", c.Prefix, err)
					}
					if e.ExitStatus() != 15 {
						log.Fatalf("%sexit %v", c.Prefix, e.ExitStatus())
					}
				}
			}

		}
	}

	//TODO: We should wait for all io.Copy() goroutines.
	//TODO: We should not exit 0, if there was an error.
}
