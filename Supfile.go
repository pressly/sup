package stackup

// Config represents the configuration data that are
// loaded from the Supfile YAML file.
type Config struct {
	Networks map[string]Network  `yaml:"networks"`
	Commands map[string]Command  `yaml:"commands"`
	Targets  map[string][]string `yaml:"targets"`
	Env      map[string]string   `yaml:"env"`
}

// Network represents the group of hosts with a custom env.
type Network struct {
	Hosts []string          `yaml:"hosts"`
	Env   map[string]string `yaml:"env"`
}

type Upload struct {
	Src string `yaml:"src"`
	Dst string `yaml:"dst"`
}

// Command represents set of commands to be run remotely.
type Command struct {
	Name   string   `yaml:-` // To be parsed manually.
	Desc   string   `yaml:"desc"`
	Run    string   `yaml:"run`
	Script string   `yaml:"script"` // A file to be read into Run.
	Upload []Upload `yaml:"upload"`
	Stdin  bool     `yaml:"stdin"`
}
