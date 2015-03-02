package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/pressly/prefixer"

	"gopkg.in/yaml.v2"
)

type Command struct {
	Name   string `yaml:-`
	Desc   string `yaml:"desc"`
	Exec   string `yaml:"exec`
	Script string `yaml:"script"`
	//Env map[string]string `yaml:"env"`
}

// Config represents the configuration data that are
// loaded from the Supfile YAML file.
type Config struct {
	Hosts    map[string][]string `yaml:"hosts"`
	Env      map[string]string   `yaml:"env"`
	Commands map[string]Command  `yaml:"commands"`
	Targets  map[string][]string `yaml:"targets"`
}

func usage(conf *Config) {
	switch len(os.Args) {
	case 2:
		log.Println("Usage: sup <hosts> <target/command>\n")
		fallthrough
	case 3:
		log.Println("Available targets (from Supfile):")
		for target, _ := range conf.Targets {
			log.Printf("- %v\n", target)
		}
		log.Println("Available commands (from Supfile):")
		for cmd, _ := range conf.Commands {
			log.Printf("- %v\n", cmd)
		}
	case 1:
		log.Println("Usage: sup <hosts> <target/command>")
		log.Println("Available hosts (from Supfile):")
		for group, hosts := range conf.Hosts {
			log.Printf("- %v\n", group)
			for _, host := range hosts {
				log.Printf("   - %v\n", host)
			}
		}
	}
	os.Exit(1)
}

func main() {
	var (
		conf           Config
		commands       []Command
		longestHostLen int
	)

	data, _ := ioutil.ReadFile("./Supfile")
	if err := yaml.Unmarshal(data, &conf); err != nil {
		log.Fatal(err)
	}

	if len(os.Args) != 3 {
		usage(&conf)
	}

	hosts, ok := conf.Hosts[os.Args[1]]
	if !ok || len(hosts) == 0 {
		usage(&conf)
	}

	target, isTarget := conf.Targets[os.Args[2]]
	if isTarget {
		for _, cmd := range target {
			command, isCommand := conf.Commands[cmd]
			if !isCommand {
				log.Printf("Unknown command \"%v\" (from target \"%v\": %v)\n\n", cmd, os.Args[2], target)
				usage(&conf)
			}
			command.Name = cmd
			commands = append(commands, command)
		}
	} else {
		command, isCommand := conf.Commands[os.Args[2]]
		if !isCommand {
			// Not a target, nor command
			log.Printf("Unknown target/command \"%v\"\n\n", os.Args[2])
			usage(&conf)
		}
		command.Name = os.Args[2]
		commands = append(commands, command)
	}

	clients := make([]*SSHClient, len(hosts))
	for i, host := range hosts {
		c := &SSHClient{
			Env: map[string]string{"FOO": "sup"},
		}
		if err := c.Connect(host); err != nil {
			log.Fatal(err)
		}
		defer c.Close()

		if len(c.Host) > longestHostLen {
			longestHostLen = len(c.Host)
		}

		clients[i] = c
	}

	for _, cmd := range commands {
		if cmd.Exec != "" {
			log.Printf("Run command \"%v\": Exec \"%v\"", cmd.Name, cmd.Exec)
		} else if cmd.Script != "" {
			log.Printf("Run command \"%v\": Exec script \"%v\"", cmd.Name, cmd.Script)
			f, err := os.Open(cmd.Script)
			if err != nil {
				log.Fatal(err)
			}
			data, err = ioutil.ReadAll(f)
			if err != nil {
				log.Fatal(err)
			}
			cmd.Exec = string(data)
		} else {
			log.Fatalf("Run command \"%v\": Nothing to run", cmd.Name)
		}

		for _, c := range clients {
			padding := strings.Repeat(" ", longestHostLen-len(c.Host))
			c.Prefix = padding + c.Host + " | "
			c.Run(cmd)

			go func(c *SSHClient) {
				if _, err := io.Copy(os.Stdout, prefixer.New(c.RemoteStdout, c.Prefix)); err != nil {
					log.Printf("STDOUT(%v): %v", c.Host, err)
				}
			}(c)
			go func(c *SSHClient) {
				if _, err := io.Copy(os.Stderr, prefixer.New(c.RemoteStderr, c.Prefix)); err != nil {
					log.Printf("STERR(%v): %v", c.Host, err)
				}
			}(c)
		}

		for _, c := range clients {
			_ = c.Wait()
			// TODO: check for exit err:
			// err = session.Wait()
			// if err == nil {
			// 	t.Fatalf("expected command to fail but it didn't")
			// }
			// e, ok := err.(*ExitError)
			// if !ok {
			// 	t.Fatalf("expected *ExitError but got %T", err)
			// }
			// if e.ExitStatus() != 15 {
			// 	t.Fatalf("expected command to exit with 15 but got %v", e.ExitStatus())
			// }
		}
	}
}
