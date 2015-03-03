package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/pressly/prefixer"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

// Config represents the configuration data that are
// loaded from the Supfile YAML file.
type Config struct {
	Hosts    map[string][]string `yaml:"hosts"`
	Env      map[string]string   `yaml:"env"`
	Commands map[string]Command  `yaml:"commands"`
	Targets  map[string][]string `yaml:"targets"`
}

// Command represents the
type Command struct {
	Name   string `yaml:-` // To be parsed manually.
	Desc   string `yaml:"desc"`
	Exec   string `yaml:"exec`
	Script string `yaml:"script"` // A file to be read into Exec.
}

// usage prints help for an arg and exits.
func usage(conf *Config, arg int) {
	log.Println("Usage: sup <hosts> <target/command>\n")
	switch arg {
	case 1:
		// <hosts> missing, print available hosts.
		log.Println("Available hosts (from Supfile):")
		for group, hosts := range conf.Hosts {
			log.Printf("- %v\n", group)
			for _, host := range hosts {
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

func main() {
	var (
		conf       Config
		commands   []Command
		paddingLen int
	)

	// Read configuration file.
	data, _ := ioutil.ReadFile("./Supfile")
	if err := yaml.Unmarshal(data, &conf); err != nil {
		log.Fatal(err)
	}

	// Check for the first argument first
	if len(os.Args) < 2 {
		usage(&conf, len(os.Args))
	}
	// Does the <host> exist?
	hosts, ok := conf.Hosts[os.Args[1]]
	if !ok || len(hosts) == 0 {
		usage(&conf, 1)
	}

	// Check for the second argument
	if len(os.Args) < 3 {
		usage(&conf, len(os.Args))
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
				usage(&conf, 2)
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
			usage(&conf, 2)
		}
		command.Name = os.Args[2]
		commands = append(commands, command)
	}

	// Check for extra arguments
	if len(os.Args) != 3 {
		usage(&conf, len(os.Args))
	}

	// Process all ENVs into a string of form
	// `export FOO="bar"; export BAR="baz";`.
	env := ``
	for name, value := range conf.Env {
		env += `export ` + name + `="` + value + `";`
	}

	// Open SSH connection to all the hosts.
	clients := make([]*SSHClient, len(hosts))
	for i, host := range hosts {
		c := &SSHClient{
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
		// Script? Read it into the Exec as string of commands.
		if cmd.Script != "" {
			f, err := os.Open(cmd.Script)
			if err != nil {
				log.Fatal(err)
			}
			data, err = ioutil.ReadAll(f)
			if err != nil {
				log.Fatal(err)
			}
			cmd.Exec = string(data)
		}

		// No commands specified for the command.
		if cmd.Exec == "" {
			log.Fatalf("Run command \"%v\": Nothing to run", cmd.Name)
		}

		// Run the command on all hosts in parallel.
		for i, c := range clients {
			padding := strings.Repeat(" ", paddingLen-(len(c.User)+1+len(c.Host)))

			c.Prefix = string(Colors[i%len(Colors)]) + padding + c.User + "@" + c.Host + " | "
			c.Run(cmd)

			go func(c *SSHClient) {
				if _, err := io.Copy(os.Stdout, prefixer.New(c.RemoteStdout, c.Prefix)); err != nil {
					log.Printf("%sSTDOUT error: %v", c.Prefix, c.Host, err)
				}
			}(c)
			go func(c *SSHClient) {
				if _, err := io.Copy(os.Stderr, prefixer.New(c.RemoteStderr, c.Prefix)); err != nil {
					log.Printf("%sSTDERR error: %v", c.Prefix, c.Host, err)
				}
			}(c)
		}

		// Wait for all hosts to finish.
		for _, c := range clients {
			if err := c.Wait(); err != nil {
				//TODO: Handle the SSH ExitError in ssh.go?
				e, ok := err.(*ssh.ExitError)
				if !ok {
					log.Fatalf("%sexpected *ExitError but got %T", c.Prefix, err)
				}
				if e.ExitStatus() != 15 {
					log.Fatalf("%sexit %v", c.Prefix, e.ExitStatus())
				}
			}
		}
	}

	//TODO: We should wait for all io.Copy() goroutines.
	//TODO: We should not exit 0, if there was an error.
}
