package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

var (
	conf map[string]interface{}
)

func main() {
	data, _ := ioutil.ReadFile("./Supfile")
	err := yaml.Unmarshal(data, &conf)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(conf)

	commands := conf["commands"].(map[interface{}]interface{})

	hosts := conf["hosts"].(map[interface{}]interface{})
	betaHosts := hosts["beta"].([]interface{})
	betaHost := betaHosts[0].(string)

	log.Println(betaHost)

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
