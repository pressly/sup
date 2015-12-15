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

	// Script. Read the file as a multiline input command.
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
		} else {
			task.Clients = clients
		}
		tasks = append(tasks, task)
	}

	// Local command.
	if cmd.Local != "" {
		local := &LocalhostClient{
			env: env + `export SUP_HOST="localhost";`,
		}
		local.Connect("localhost")
		task := &Task{
			Run:     cmd.Local,
			Clients: []Client{local},
		}
		if cmd.Stdin {
			task.Input = os.Stdin
		}
		tasks = append(tasks, task)
	}

	// Remote command.
	if cmd.Run != "" {
		task := Task{
			Run: cmd.Run,
		}
		if cmd.Stdin {
			task.Input = os.Stdin
		}
		if cmd.RunOnce {
			task.Clients = []Client{clients[0]}
			tasks = append(tasks, &task)
		} else if cmd.Serial > 0 {
			// Each "serial" task client group is executed sequentially.
			for i := 0; i < len(clients); i += cmd.Serial {
				j := i + cmd.Serial
				if j > len(clients) {
					j = len(clients)
				}
				copy := task
				copy.Clients = clients[i:j]
				tasks = append(tasks, &copy)
			}
		} else {
			task.Clients = clients
			tasks = append(tasks, &task)
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
