package sshconfig

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// SSHHost defines a single host entry in a ssh config
type SSHHost struct {
	Host              []string
	HostName          string
	User              string
	Port              int
	ProxyCommand      string
	HostKeyAlgorithms string
	IdentityFile      string
}

// MustParseSSHConfig must parse the SSH config given by path or it will panic
func MustParseSSHConfig(path string) []*SSHHost {
	config, err := ParseSSHConfig(path)
	if err != nil {
		panic(err)
	}
	return config
}

// ParseSSHConfig parses a SSH config given by path.
func ParseSSHConfig(path string) ([]*SSHHost, error) {
	// read config file
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parse(string(content))
}

// parses an openssh config file
func parse(input string) ([]*SSHHost, error) {
	sshConfigs := []*SSHHost{}
	var next item
	var sshHost *SSHHost

	lexer := lex(input)
Loop:
	for {
		token := lexer.nextItem()

		if sshHost == nil && token.typ != itemHost {
			return nil, fmt.Errorf("config variable before Host variable")
		}

		switch token.typ {
		case itemHost:
			if sshHost != nil {
				sshConfigs = append(sshConfigs, sshHost)
			}

			sshHost = &SSHHost{Host: []string{}, Port: 22}
		case itemHostValue:
			sshHost.Host = strings.Split(token.val, " ")
		case itemHostName:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			sshHost.HostName = next.val
		case itemUser:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			sshHost.User = next.val
		case itemPort:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			port, err := strconv.Atoi(next.val)
			if err != nil {
				return nil, err
			}
			sshHost.Port = port
		case itemProxyCommand:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			sshHost.ProxyCommand = next.val
		case itemHostKeyAlgorithms:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			sshHost.HostKeyAlgorithms = next.val
		case itemIdentityFile:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			sshHost.IdentityFile = next.val
		case itemError:
			return nil, fmt.Errorf("%s at pos %d", token.val, token.pos)
		case itemEOF:
			if sshHost != nil {
				sshConfigs = append(sshConfigs, sshHost)
			}
			break Loop
		default:
			// continue onwards
		}
	}
	return sshConfigs, nil
}
