package sup

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Supfile struct {
	Networks Networks `yaml:"networks"`
	Commands Commands `yaml:"commands"`
	Targets  Targets  `yaml:"targets"`
	Env      EnvList  `yaml:"env"`
	Version  string   `yaml:"version"`
}

type Network struct {
	Env          EnvList  `yaml:"env"`
	Inventory    string   `yaml:"inventory"`
	Hosts        []string `yaml:"hosts"`
	Bastion      string   `yaml:"bastion"`
	Password     string   `yaml:"password"`
	User         string   `yaml:"user"`
	IdentityFile string   `yaml:"identity_file"`
}

type Networks struct {
	Names []string
	nets  map[string]Network
}

func (n *Networks) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(&n.nets)
	if err != nil {
		return err
	}

	var items yaml.MapSlice
	err = unmarshal(&items)
	if err != nil {
		return err
	}

	n.Names = make([]string, len(items))
	for i, item := range items {
		n.Names[i] = item.Key.(string)
	}

	return nil
}

func (n *Networks) Get(name string) (Network, bool) {
	net, ok := n.nets[name]
	return net, ok
}

type Command struct {
	Name   string   `yaml:"-"`
	Desc   string   `yaml:"desc"`
	Local  string   `yaml:"local"`
	Run    string   `yaml:"run"`
	Script string   `yaml:"script"`
	Upload []Upload `yaml:"upload"`
	Stdin  bool     `yaml:"stdin"`
	Once   bool     `yaml:"once"`
	Serial int      `yaml:"serial"`

	RunOnce bool `yaml:"run_once"`
}

type Commands struct {
	Names []string
	cmds  map[string]Command
}

func (c *Commands) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(&c.cmds)
	if err != nil {
		return err
	}

	var items yaml.MapSlice
	err = unmarshal(&items)
	if err != nil {
		return err
	}

	c.Names = make([]string, len(items))
	for i, item := range items {
		c.Names[i] = item.Key.(string)
	}

	return nil
}

func (c *Commands) Get(name string) (Command, bool) {
	cmd, ok := c.cmds[name]
	return cmd, ok
}

type Targets struct {
	Names   []string
	targets map[string][]string
}

func (t *Targets) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(&t.targets)
	if err != nil {
		return err
	}

	var items yaml.MapSlice
	err = unmarshal(&items)
	if err != nil {
		return err
	}

	t.Names = make([]string, len(items))
	for i, item := range items {
		t.Names[i] = item.Key.(string)
	}

	return nil
}

func (t *Targets) Get(name string) ([]string, bool) {
	cmds, ok := t.targets[name]
	return cmds, ok
}

type Upload struct {
	Src     string `yaml:"src"`
	Dst     string `yaml:"dst"`
	Exc     string `yaml:"exclude"`
	Flatten bool   `yaml:"flatten"`
	Sudo    bool   `yaml:"sudo"`
}

type EnvVar struct {
	Key   string
	Value string
}

func (e EnvVar) String() string {
	return e.Key + `=` + e.Value
}

func (e EnvVar) AsExport() string {
	return `export ` + e.Key + `="` + e.Value + `";`
}

type EnvList []*EnvVar

func (e EnvList) Slice() []string {
	envs := make([]string, len(e))
	for i, env := range e {
		envs[i] = env.String()
	}
	return envs
}

func (e *EnvList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	items := []yaml.MapItem{}

	err := unmarshal(&items)
	if err != nil {
		return err
	}

	*e = make(EnvList, 0, len(items))

	for _, v := range items {
		e.Set(fmt.Sprintf("%v", v.Key), fmt.Sprintf("%v", v.Value))
	}

	return nil
}

func (e *EnvList) Set(key, value string) {
	for i, v := range *e {
		if v.Key == key {
			(*e)[i].Value = value
			return
		}
	}

	*e = append(*e, &EnvVar{
		Key:   key,
		Value: value,
	})
}

func (e *EnvList) ResolveValues() error {
	if len(*e) == 0 {
		return nil
	}

	exports := ""
	for i, v := range *e {
		exports += v.AsExport()

		cmd := exec.Command("bash", "-c", exports+"echo -n "+v.Value+";")
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		cmd.Dir = cwd
		resolvedValue, err := cmd.Output()
		if err != nil {
			return errors.Wrapf(err, "resolving env var %v failed", v.Key)
		}

		(*e)[i].Value = string(resolvedValue)
	}

	return nil
}

func (e *EnvList) AsExport() string {
	exports := ``
	for _, v := range *e {
		exports += v.AsExport() + " "
	}
	return exports
}

type ErrMustUpdate struct {
	Msg string
}

type ErrUnsupportedSupfileVersion struct {
	Msg string
}

func (e ErrMustUpdate) Error() string {
	return fmt.Sprintf("%v\n\nPlease update sup by `go get -u github.com/pressly/sup/cmd/sup`", e.Msg)
}

func (e ErrUnsupportedSupfileVersion) Error() string {
	return fmt.Sprintf("%v\n\nCheck your Supfile version (available latest version: v0.5)", e.Msg)
}

func NewSupfile(data []byte) (*Supfile, error) {
	var conf Supfile

	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	switch conf.Version {
	case "":
		conf.Version = "0.1"
		fallthrough

	case "0.1":
		for _, cmd := range conf.Commands.cmds {
			if cmd.RunOnce {
				return nil, ErrMustUpdate{"command.run_once is not supported in Supfile v" + conf.Version}
			}
		}
		fallthrough

	case "0.2":
		for _, cmd := range conf.Commands.cmds {
			if cmd.Once {
				return nil, ErrMustUpdate{"command.once is not supported in Supfile v" + conf.Version}
			}
			if cmd.Local != "" {
				return nil, ErrMustUpdate{"command.local is not supported in Supfile v" + conf.Version}
			}
			if cmd.Serial != 0 {
				return nil, ErrMustUpdate{"command.serial is not supported in Supfile v" + conf.Version}
			}
		}
		for _, network := range conf.Networks.nets {
			if network.Inventory != "" {
				return nil, ErrMustUpdate{"network.inventory is not supported in Supfile v" + conf.Version}
			}
		}
		fallthrough

	case "0.3":
		var warning string
		for key, cmd := range conf.Commands.cmds {
			if cmd.RunOnce {
				warning = "Warning: command.run_once was deprecated by command.once in Supfile v" + conf.Version + "\n"
				cmd.Once = true
				conf.Commands.cmds[key] = cmd
			}
		}
		if warning != "" {
			fmt.Fprintf(os.Stderr, warning)
		}

		fallthrough

	case "0.4", "0.5", "0.5.1", "0.5.2":
	default:
		return nil, ErrUnsupportedSupfileVersion{"unsupported Supfile version " + conf.Version}
	}

	return &conf, nil
}

func (n Network) ParseInventory() ([]string, error) {
	if n.Inventory == "" {
		return nil, nil
	}

	cmd := exec.Command("/bin/sh", "-c", n.Inventory)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, n.Env.Slice()...)
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
		if host == "" || host[:1] == "#" {
			continue
		}

		hosts = append(hosts, host)
	}
	return hosts, nil
}
