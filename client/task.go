package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pressly/stackup/config"
)

// Task represents a set of tasks to be run.
// TODO(VojtechVitek): This is just blindly copy/pasted from config.Command.
type Task struct {
	Name string
	Desc string
	Run  string
}

func TasksFromConfigCommand(cmd config.Command) ([]Task, error) {
	var tasks []Task

	// Script? Read it into the Run as string of commands.
	if cmd.Script != "" {
		f, err := os.Open(cmd.Script)
		if err != nil {
			log.Fatal(err)
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatal(err)
		}
		task := Task{
			Run: string(data),
		}
		tasks = append(tasks, task)
	}

	// No commands specified for the command.
	if cmd.Run != "" {
		task := Task{
			Run: cmd.Run,
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

type ErrTask struct {
	Task   Task
	Reason string
}

func (e ErrTask) Error() string {
	return fmt.Sprintf(`Run("%v"): %v`, e.Task, e.Reason)
}
