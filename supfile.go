package sup

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

// Supfile represents the Stackup configuration YAML file.
type Supfile struct {
	Networks map[string]Network  `yaml:"networks"`
	Commands map[string]Command  `yaml:"commands"`
	Targets  map[string][]string `yaml:"targets"`
	Env      map[string]string   `yaml:"env"`
	Version  string              `yaml:"version"`
}

// Network is group of hosts with extra custom env vars.
type Network struct {
	Env       map[string]string `yaml:"env"`
	Inventory string            `yaml:"inventory"`
	Hosts     []string          `yaml:"hosts"`
	Bastion   string            `yaml:"bastion"` // Jump host for the environment
}

// Command represents command(s) to be run remotely.
type Command struct {
	Name    string   `yaml:"-"`        // Command name.
	Desc    string   `yaml:"desc"`     // Command description.
	Run     string   `yaml:"run"`      // Command(s) to be run remotelly.
	Script  string   `yaml:"script"`   // Load command(s) from script and run it remotelly.
	Upload  []Upload `yaml:"upload"`   // See below.
	Stdin   bool     `yaml:"stdin"`    // Attach localhost STDOUT to remote commands' STDIN?
	Max     int      `yaml:"max"`      // Max number of clients processing a task in parallel.
	RunOnce bool     `yaml:"run_once"` // The command should be run once only.
	// TODO: RunSerial int      `yaml:"run_serial"` // Max number of clients processing the command in parallel.
}

// Upload represents file copy operation from localhost Src path to Dst
// path of every host in a given Network.
type Upload struct {
	Src string `yaml:"src"`
	Dst string `yaml:"dst"`
	Exc string `yaml:"exclude"`
}

// NewSupfile parses configuration file and returns Supfile or error.
func NewSupfile(file string) (*Supfile, error) {
	var conf Supfile
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}

	switch conf.Version {
	case "", "0.1":
		for _, cmd := range conf.Commands {
			if cmd.RunOnce {
				return nil, errors.New("command.run_once is not supported in Supfile version 0.1")
			}
		}
	case "0.2":
		// latest; skip
	default:
		return nil, errors.New("unsupported version, please update sup by `go get -u github.com/pressly/sup`")
	}

	for i, network := range conf.Networks {
		hosts, err := network.ParseInventory()
		if err != nil {
			return nil, err
		}
		network.Hosts = append(network.Hosts, hosts...)
		conf.Networks[i] = network
	}

	return &conf, nil
}

// ParseInventory runs the inventory command, if provided, and appends
// the command's output lines to the manually defined list of hosts.
func (n Network) ParseInventory() ([]string, error) {
	if n.Inventory == "" {
		return nil, nil
	}

	cmd := exec.Command("/bin/sh", "-c", n.Inventory)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var hosts []string
	buf := bytes.NewBuffer(output)
	for {
		host, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		host = strings.TrimSpace(host)
		// skip empty lines and comments
		if host == "" || host[:1] == "#" {
			continue
		}

		hosts = append(hosts, host)
	}
	return hosts, nil
}
