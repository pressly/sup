package stackup

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// Task represents a set of commands to be run.
type Task struct {
	Name string
	Run  string
}

func TasksFromConfigCommand(cmd Command) ([]Task, error) {
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
