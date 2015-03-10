package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pressly/stackup"

	"github.com/pressly/prefixer"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

// usage prints help for an arg and exits.
func usage(conf *stackup.Config, arg int) {
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
		for name, commands := range conf.Targets {
			log.Printf("- %v", name)
			for _, cmd := range commands {
				log.Printf("\t%v\n", cmd)
			}
		}
		log.Println()
		log.Println("Available commands (from Supfile):")
		for name, cmd := range conf.Commands {
			log.Printf("- %v\t%v", name, cmd.Desc)
		}
	}
	os.Exit(1)
}

// parseArgs parses os.Args for network and commands to be run.
func parseArgsOrDie(conf *stackup.Config) (stackup.Network, []stackup.Command) {
	var commands []stackup.Command

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

// parseConfigFileOrDie reads and parses configuration from Supfile in CWD.
func parseConfigFileOrDie(file string) *stackup.Config {
	var conf stackup.Config
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		log.Fatal(err)
	}
	return &conf
}

func main() {
	// Parse configuration.
	// TODO: -f flag.
	conf := parseConfigFileOrDie("./Supfile")

	// Parse network and commands to be run from os.Args.
	network, commands := parseArgsOrDie(conf)

	// Process all ENVs into a string of form
	// `export FOO="bar"; export BAR="baz";`.
	env := ``
	for name, value := range conf.Env {
		env += `export ` + name + `="` + value + `";`
	}
	for name, value := range network.Env {
		env += `export ` + name + `="` + value + `";`
	}

	var paddingLen int

	// Create clients for every host (either SSH or Localhost).
	var clients []stackup.Client
	for _, host := range network.Hosts {
		var c stackup.Client

		if host == "localhost" { // LocalhostClient

			local := &stackup.LocalhostClient{
				Env: env,
			}
			if err := local.Connect(host); err != nil {
				log.Fatal(err)
			}

			c = local

		} else { // SSHClient

			remote := &stackup.SSHClient{
				Env: env,
			}
			if err := remote.Connect(host); err != nil {
				log.Fatal(err)
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

	// Run command or run multiple commands defined by target sequentally.
	for _, cmd := range commands {
		// Translate command into task(s).
		tasks, err := stackup.TasksFromConfigCommand(cmd, env)
		if err != nil {
			log.Fatalf("TasksFromConfigCommand(): ", err)
		}

		// Run tasks sequentally.
		for _, task := range tasks {

			var writers []io.Writer

			// Run task in parallel.
			for i, c := range clients {
				padding := strings.Repeat(" ", paddingLen-(len(c.Prefix())))
				color := stackup.Colors[i%len(stackup.Colors)]

				prefix := color + padding + c.Prefix() + " | "
				err := c.Run(task)
				if err != nil {
					log.Fatalf("%sexit %v", prefix, err)
				}

				// Copy over tasks's STDOUT.
				go func(c stackup.Client) {
					switch t := c.(type) {
					case *stackup.SSHClient:
						if _, err := io.Copy(os.Stdout, prefixer.New(t.RemoteStdout, prefix)); err != nil {
							log.Printf("%sSTDOUT: %v", t.Prefix(), err)
						}
					case *stackup.LocalhostClient:
						if _, err := io.Copy(os.Stdout, prefixer.New(t.Stdout, prefix)); err != nil {
							log.Printf("%sSTDOUT: %v", t.Prefix(), err)
						}
					}
				}(c)

				// Copy over tasks's STDERR.
				go func(c stackup.Client) {
					switch t := c.(type) {
					case *stackup.SSHClient:
						if _, err := io.Copy(os.Stderr, prefixer.New(t.RemoteStderr, prefix)); err != nil {
							log.Printf("%sSTDERR: %v", t.Prefix(), err)
						}
					case *stackup.LocalhostClient:
						if _, err := io.Copy(os.Stderr, prefixer.New(t.Stderr, prefix)); err != nil {
							log.Printf("%sSTDERR: %v", t.Prefix(), err)
						}
					}
				}(c)

				switch t := c.(type) {
				case *stackup.SSHClient:
					writers = append(writers, t.RemoteStdin)
				case *stackup.LocalhostClient:
					writers = append(writers, t.Stdin)
				}
			}

			// Copy over task's STDIN.
			if task.Input != nil {
				writer := io.MultiWriter(writers...)
				_, err := io.Copy(writer, task.Input)
				if err != nil {
					log.Printf("STDIN: %v", err)
				}
			}

			//TODO: Use MultiWriteCloser (not in Stdlib), so we can writer.Close()?
			// 	    Move this at least to some defer function instead.
			for _, c := range clients {
				c.WriteClose()
			}

			// Wait for all clients to finish the task.
			for _, c := range clients {
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
			}
		}
	}

	//TODO: We should wait for all io.Copy() goroutines.
	//Ugly hack for now:
	time.Sleep(1000 * time.Millisecond)
}
