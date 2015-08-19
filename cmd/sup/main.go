package main

import (
	"log"
	"os"

	"github.com/pressly/sup"
)

// usage prints help for an arg and exits.
func usage(conf *sup.Supfile, arg int) {
	log.Println("Usage: sup <network> <target/command>\n")
	switch arg {
	case 1:
		// <network> missing, print available hosts.
		log.Println("Available networks (from ./Supfile):")
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

// parseArgs parses os.Args and returns network and commands to be run.
// On error, it prints usage and exits.
func parseArgsOrDie(conf *sup.Supfile) (*sup.Network, []*sup.Command) {
	var commands []*sup.Command

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
			commands = append(commands, &command)
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
		commands = append(commands, &command)
	}

	// Check for extra arguments
	if len(os.Args) != 3 {
		usage(conf, len(os.Args))
	}

	return &network, commands
}

func main() {
	// Parse configuration file in current directory.
	// TODO: -f flag to pass custom file.
	conf, err := sup.NewSupfile("./Supfile")
	if err != nil {
		log.Fatal(err)
	}

	// Parse network and commands to be run from os.Args.
	network, commands := parseArgsOrDie(conf)

	// Create new Stackup app.
	app, err := sup.New(conf)
	if err != nil {
		log.Fatal(err)
	}

	// Run all the commands in the given network.
	err = app.Run(network, commands...)
	if err != nil {
		log.Fatal(err)
	}
}
