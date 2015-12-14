package sup

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// Task represents a set of commands to be run.
type Task struct {
	Run     string
	Input   io.Reader
	Clients []Client
}

func CreateTasks(cmd *Command, clients []Client, env string) ([]*Task, error) {
	var tasks []*Task

	// Anything to upload?
	for _, upload := range cmd.Upload {
		task := &Task{
			Run:   RemoteTarCommand(upload.Dst),
			Input: NewTarStreamReader(upload.Src, upload.Exc, env),
		}

		if cmd.RunOnce {
			task.Clients = []Client{clients[0]}
			tasks = append(tasks, task)
		} else {
			task.Clients = clients
			tasks = append(tasks, task)
		}
	}

	// Script? Read the file as a multiline input command.
	if cmd.Script != "" {
		f, err := os.Open(cmd.Script)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		task := &Task{
			Run: string(data),
		}
		if cmd.Stdin {
			task.Input = os.Stdin
		}

		if cmd.RunOnce {
			task.Clients = []Client{clients[0]}
			tasks = append(tasks, task)
		} else {
			task.Clients = clients
			tasks = append(tasks, task)
		}
	}

	// Command?
	if cmd.Run != "" {
		task := &Task{
			Run: cmd.Run,
		}
		if cmd.Stdin {
			task.Input = os.Stdin
		}

		if cmd.RunOnce {
			task.Clients = []Client{clients[0]}
			tasks = append(tasks, task)
		} else {
			task.Clients = clients
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

type ErrTask struct {
	Task   *Task
	Reason string
}

func (e ErrTask) Error() string {
	return fmt.Sprintf(`Run("%v"): %v`, e.Task, e.Reason)
}
