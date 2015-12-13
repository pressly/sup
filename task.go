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
	RunOnce bool
	// TODO: RunSerial int
}

func TasksFromConfigCommand(cmd *Command, env string) ([]*Task, error) {
	var tasks []*Task

	// Anything to upload?
	for _, upload := range cmd.Upload {
		task := &Task{
			Run:   RemoteTarCommand(upload.Dst),
			Input: NewTarStreamReader(upload.Src, upload.Exc, env),
		}

		tasks = append(tasks, task)
	}

	// Script? Read it as a set of commands.
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
			Run:     string(data),
			RunOnce: cmd.RunOnce,
			// TODO: RunSerial: cmd.RunSerial,
		}
		if cmd.Stdin {
			task.Input = os.Stdin
		}

		tasks = append(tasks, task)
	}

	// Command?
	if cmd.Run != "" {
		task := &Task{
			Run:     cmd.Run,
			RunOnce: cmd.RunOnce,
			// TODO: RunSerial: cmd.RunSerial,
		}
		if cmd.Stdin {
			task.Input = os.Stdin
		}

		tasks = append(tasks, task)
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
