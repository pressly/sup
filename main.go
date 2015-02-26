package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Command struct {
	Desc   string `yaml:"desc"`
	Exec   string `yaml:"exec`
	Script string `yaml:"script"`
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
		log.Fatal("Available commands (from Supfile):\n")
		for cmd, _ := range conf.Commands {
			log.Fatal("- %v\n", cmd)
		}
	case 1:
		log.Fatal("Available hosts (from Supfile):\n")
		for group, hosts := range conf.Hosts {
			log.Fatal("- %v\n", group)
			for _, host := range hosts {
				log.Fatal("   - %v\n", host)
			}
		}
	default:
		log.Fatal("Usage:\nsup <host-group> <command-alias>")
	}
	os.Exit(1)
}

func main() {
	var conf Config
	data, _ := ioutil.ReadFile("./Supfile")
	if err := yaml.Unmarshal(data, &conf); err != nil {
		log.Fatal(err)
	}

	if len(os.Args) != 3 {
		usage(&conf)
	}

	hosts, ok := conf.Hosts[os.Args[1]]
	if !ok {
		usage(&conf)
	}

	// TODO: Targets?

	command, ok := conf.Commands[os.Args[2]]
	if !ok {
		usage(&conf)
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
		clients[i] = c
	}

	for _, c := range clients {
		c.Run(command)

		go func(c *SSHClient) {
			if _, err := io.Copy(os.Stdout, c.RemoteStdout); err != nil {
				log.Printf("STDOUT(%v): %v", c.Host, err)
			}
			if _, err := io.Copy(os.Stderr, c.RemoteStderr); err != nil {
				log.Printf("STERR(%v): %v", c.Host, err)
			}
		}(c)

	}

	for _, c := range clients {
		c.Wait()
	}

	// 1. Parse the command that we're going to run..
	// ie. from:
	// $ sup beta deploy  ===> 1. connect to beta  2. run deploy command
	// $ sup beta stop
	// $ sup beta start
	// command := commands["ping"].(map[interface{}]interface{}) // ..
	// cmd := command["exec"].(string)

	// log.Printf("the command: %s", cmd)

	// log.Println(b)
	// s.Session.Wait()

	// s.Run(cmd)
	// s.Quit()

	// 2. Open connection to each host - all must connect

	// 3.

	//s.Session.Wait()
}
