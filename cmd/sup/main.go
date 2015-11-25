package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/pressly/sup"
)

var (
	supfile          = flag.String("f", "./Supfile", "custom path to Supfile")
	showVersionShort = flag.Bool("v", false, "print version")
	showVersionLong  = flag.Bool("version", false, "print version")

	ErrCmd              = errors.New("Usage: sup [-f <Supfile>] <network> <target/command>")
	ErrUnknownNetwork   = errors.New("Unknown network")
	ErrNetworkNoHosts   = errors.New("No hosts for a given network")
	ErrTarget           = errors.New("Unknown target")
	ErrTargetNoCommands = errors.New("No commands for a given target")
)

func networkUsage(conf *sup.Supfile) {
	w := &tabwriter.Writer{}
	w.Init(os.Stderr, 4, 4, 2, ' ', 0)
	defer w.Flush()

	// <network> missing, print available hosts.
	fmt.Fprintln(w, "Networks:\t")
	for name, network := range conf.Networks {
		fmt.Fprintf(w, "- %v\n", name)
		for _, host := range network.Hosts {
			fmt.Fprintf(w, "\t- %v\n", host)
		}
	}
	fmt.Fprintln(w)
}

func targetUsage(conf *sup.Supfile) {
	w := &tabwriter.Writer{}
	w.Init(os.Stderr, 4, 4, 2, ' ', 0)
	defer w.Flush()

	// <target/command> not found or missing,
	// print available targets/commands.
	fmt.Fprintln(w, "Targets:\t")
	for name, commands := range conf.Targets {
		fmt.Fprintf(w, "- %v\t%v\n", name, strings.Join(commands, ", "))
	}
	fmt.Fprintln(w, "\t")
	fmt.Fprintln(w, "Commands:\t")
	for name, cmd := range conf.Commands {
		fmt.Fprintf(w, "- %v\t%v\n", name, cmd.Desc)
	}
	fmt.Fprintln(w)
}

// parseArgs parses args and returns network and commands to be run.
// On error, it prints usage and exits.
func parseArgs(conf *sup.Supfile) (*sup.Network, []*sup.Command, error) {
	var commands []*sup.Command

	args := flag.Args()

	if len(args) < 1 {
		networkUsage(conf)
		return nil, nil, ErrCmd
	}

	// Does the <network> exist?
	network, ok := conf.Networks[args[0]]
	if !ok {
		networkUsage(conf)
		return nil, nil, ErrUnknownNetwork
	}

	// Does <network> have any hosts?
	if len(network.Hosts) == 0 {
		networkUsage(conf)
		return nil, nil, ErrNetworkNoHosts
	}

	// Check for the second argument
	if len(args) < 2 {
		targetUsage(conf)
		return nil, nil, ErrCmd
	}

	// Does the <target/command> exist?
	target, isTarget := conf.Targets[args[1]]
	if isTarget {
		// It's the target. Loop over its commands.
		for _, cmd := range target {
			// Does the target's command exist?
			command, isCommand := conf.Commands[cmd]
			if !isCommand {
				targetUsage(conf)
				return nil, nil, ErrTargetNoCommands
			}
			command.Name = cmd
			commands = append(commands, &command)
		}
	} else {
		// It's probably a command. Does it exist?
		command, isCommand := conf.Commands[args[1]]
		if !isCommand {
			// Not a target, nor command.
			targetUsage(conf)
			return nil, nil, ErrTargetNoCommands
		}
		command.Name = args[1]
		commands = append(commands, &command)
	}

	// TODO: Do we want to use extra args?
	if len(args) > 2 {
		return nil, nil, ErrCmd
	}

	return &network, commands, nil
}

func main() {
	flag.Parse()

	if *showVersionShort || *showVersionLong {
		fmt.Println("0.2")
		return
	}

	conf, err := sup.NewSupfile(*supfile)
	if err != nil {
		log.Fatal(err)
	}

	// Parse network and commands to be run from args.
	network, commands, err := parseArgs(conf)
	if err != nil {
		log.Fatal(err)
	}

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
