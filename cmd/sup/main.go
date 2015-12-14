package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/pressly/sup"
)

var (
	supfile          = flag.String("f", "./Supfile", "custom path to Supfile")
	showVersionShort = flag.Bool("v", false, "print version")
	showVersionLong  = flag.Bool("version", false, "print version")
	onlyHosts        = flag.String("only", "", "filter hosts with regexp")

	ErrUsage            = errors.New("Usage: sup [-f <Supfile>] [--only host1] <network> <target/command> [...]")
	ErrUnknownNetwork   = errors.New("Unknown network")
	ErrNetworkNoHosts   = errors.New("No hosts defined for a given network")
	ErrCmd              = errors.New("Unknown command/target")
	ErrTargetNoCommands = errors.New("No commands defined for a given target")
)

func networkUsage(conf *sup.Supfile) {
	w := &tabwriter.Writer{}
	w.Init(os.Stderr, 4, 4, 2, ' ', 0)
	defer w.Flush()

	// Print available networks/hosts.
	fmt.Fprintln(w, "Networks:\t")
	for name, network := range conf.Networks {
		fmt.Fprintf(w, "- %v\n", name)
		for _, host := range network.Hosts {
			fmt.Fprintf(w, "\t- %v\n", host)
		}
	}
	fmt.Fprintln(w)
}

func cmdUsage(conf *sup.Supfile) {
	w := &tabwriter.Writer{}
	w.Init(os.Stderr, 4, 4, 2, ' ', 0)
	defer w.Flush()

	// Print available targets/commands.
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
		return nil, nil, ErrUsage
	}

	// Does the <network> exist?
	network, ok := conf.Networks[args[0]]
	if !ok {
		networkUsage(conf)
		return nil, nil, ErrUnknownNetwork
	}

	// Does the <network> have at least one host?
	if len(network.Hosts) == 0 {
		networkUsage(conf)
		return nil, nil, ErrNetworkNoHosts
	}

	// Check for the second argument
	if len(args) < 2 {
		cmdUsage(conf)
		return nil, nil, ErrUsage
	}

	// In case of the network.Env needs an initialization
	if network.Env == nil {
		network.Env = make(map[string]string)
	}

	// Add default env variable with current network
	network.Env["SUP_NETWORK"] = args[0]

	for _, cmd := range args[1:] {
		// Target?
		target, isTarget := conf.Targets[cmd]
		if isTarget {
			// Loop over target's commands.
			for _, cmd := range target {
				command, isCommand := conf.Commands[cmd]
				if !isCommand {
					cmdUsage(conf)
					return nil, nil, fmt.Errorf("%v: %v", ErrCmd, cmd)
				}
				command.Name = cmd
				commands = append(commands, &command)
			}
		}

		// Command?
		command, isCommand := conf.Commands[cmd]
		if isCommand {
			command.Name = cmd
			commands = append(commands, &command)
		}

		if !isTarget && !isCommand {
			cmdUsage(conf)
			return nil, nil, fmt.Errorf("%v: %v", ErrCmd, cmd)
		}
	}

	return &network, commands, nil
}

func main() {
	flag.Parse()

	if *showVersionShort || *showVersionLong {
		fmt.Println(sup.VERSION)
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

	// --only option to filter hosts
	if *onlyHosts != "" {
		expr, err := regexp.CompilePOSIX(*onlyHosts)
		if err != nil {
			log.Fatal(err)
		}

		var hosts []string
		for _, host := range network.Hosts {
			if expr.MatchString(host) {
				hosts = append(hosts, host)
			}
		}
		if len(hosts) == 0 {
			log.Fatal(fmt.Errorf("no hosts match '%v' regexp", *onlyHosts))
		}
		network.Hosts = hosts
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
