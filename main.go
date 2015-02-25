package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Hosts    map[string][]string `yaml:"hosts"`
	Env      map[string]string   `yaml:"env"`
	Commands map[interface{}]struct {
		Desc    string   `yaml:"desc"`
		Exec    string   `yaml:"exec`
		Script  string   `yaml:"script"`
		Targets []string `yaml:"targets"`
	} `yaml:"commands"`
}

var (
	conf Config
)

func main() {
	data, _ := ioutil.ReadFile("./Supfile")
	err := yaml.Unmarshal(data, &conf)
	if err != nil {
		log.Fatal(err)
	}

	// log.Printf("Hosts:\n")
	// for group, hosts := range conf.Hosts {
	// 	log.Printf("- %v\n", group)
	// 	for _, host := range hosts {
	// 		log.Printf("   - %v\n", host)
	// 	}
	// }

	// log.Printf("Env:\n")
	// for name, value := range conf.Env {
	// 	log.Printf("- %v=\"%v\"\n", name, value)
	// }

	// log.Printf("Commands:\n")
	// for alias, cmd := range conf.Commands {
	// 	log.Printf("- %v\n", alias)
	// 	log.Printf("  - Desc: %v\n", cmd.Desc)
	// 	log.Printf("  - Exec: %v\n", cmd.Exec)
	// 	log.Printf("  - Script: %v\n", cmd.Script)
	// 	log.Printf("  - Targets:\n")
	// 	for _, target := range cmd.Targets {
	// 		log.Printf("    - %v\n", target)
	// 	}
	// }

	s := &SSHClient{}
	err = s.Connect(betaHost)
	if err != nil {
		log.Fatal(err)
	}

	// 1. Parse the command that we're going to run..
	// ie. from:
	// $ sup beta deploy  ===> 1. connect to beta  2. run deploy command
	// $ sup beta stop
	// $ sup beta start
	command := commands["ping"].(map[interface{}]interface{}) // ..
	cmd := command["exec"].(string)

	log.Printf("the command: %s", cmd)

	// log.Println(b)
	// s.Session.Wait()

	// s.Run(cmd)
	// s.Quit()

	// 2. Open connection to each host - all must connect

	// 3.

	s.Session.Wait()
}
