package main

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/mikkeloscar/sshconfig"
	"github.com/pressly/sup"
)

func TestSSH(t *testing.T) {
	outputs, err := setupMockEnv(sshConfigFilename, 3)
	if err != nil {
		t.Fatal(err)
	}

	flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
	flag.CommandLine.Parse([]string{"local", "t"})

	input := `
---
version: 0.4

networks:
  local:
    hosts:
    - server0
    - server1
    - server2

commands:
  test:
    run: echo "Hey over there"
  test2:
    run: echo "Hey again"

targets:
  t:
  - test
  - test2
`
	conf, err := sup.NewSupfile([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	network, commands, err := parseArgs(conf)
	if err != nil {
		t.Fatal(err)
	}

	confHosts, err := sshconfig.ParseSSHConfig(sshConfigFilename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// flatten Host -> *SSHHost, not the prettiest
	// but will do
	confMap := map[string]*sshconfig.SSHHost{}
	for _, conf := range confHosts {
		for _, host := range conf.Host {
			confMap[host] = conf
		}
	}

	// check network.Hosts for match
	for i, host := range network.Hosts {
		conf, found := confMap[host]
		if found {
			network.User = conf.User
			network.IdentityFile = resolvePath(conf.IdentityFile)
			network.Hosts[i] = fmt.Sprintf("%s:%d", conf.HostName, conf.Port)
		}
	}

	var vars sup.EnvList
	for _, val := range append(conf.Env, network.Env...) {
		vars.Set(val.Key, val.Value)
	}
	if err := vars.ResolveValues(); err != nil {
		t.Fatal(err)
	}

	app, err := sup.New(conf)
	if err != nil {
		t.Fatal(err)
	}

	err = app.Run(network, vars, commands...)
	if err != nil {
		t.Fatal(err)
	}
	for i, output := range outputs {
		t.Log(i, output.String())
	}
}
