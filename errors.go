package main

import "fmt"

type ErrConnect struct {
	User   string
	Host   string
	Reason string
}

func (e ErrConnect) Error() string {
	return fmt.Sprintf(`Connect("%v@%v"): %v`, e.User, e.Host, e.Reason)
}

type ErrCmd struct {
	Cmd    Command
	Reason string
}

func (e ErrCmd) Error() string {
	return fmt.Sprintf(`Run("%v"): %v`, e.Cmd, e.Reason)
}
