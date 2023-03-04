package sup

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

// Task represents a set of commands to be run.
type Task struct {
	Run     string
	Input   io.Reader
	Clients []Client
	TTY     bool
}

func (sup *Stackup) createTasks(cmd *Command, clients []Client, env string) ([]*Task, error) {
	var tasks []*Task

	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "resolving CWD failed")
	}

	// Anything to upload?
	for _, upload := range cmd.Upload {
		uploadFile, err := ResolveLocalPath(cwd, upload.Src, env)
		if err != nil {
			return nil, errors.Wrap(err, "upload: "+upload.Src)
		}

		uploadTarReader, err := NewTarStreamReader(cwd, uploadFile, upload.Exc)
		if err != nil {
			return nil, errors.Wrap(err, "upload: "+upload.Src)
		}

		task := Task{
			Run:   RemoteTarCommand(upload.Dst),
			Input: uploadTarReader,
			TTY:   false,
		}

		switch {
		case cmd.Once:
			task.Clients = []Client{clients[0]}
			tasks = append(tasks, &task)
		case cmd.Serial > 0:
			// Each "serial" task client group is executed sequentially.
			for i := 0; i < len(clients); i += cmd.Serial {
				j := i + cmd.Serial
				if j > len(clients) {
					j = len(clients)
				}

				taskCopy := task
				taskCopy.Clients = clients[i:j]
				tasks = append(tasks, &taskCopy)
			}
		default:
			task.Clients = clients
			tasks = append(tasks, &task)
		}
	}

	// Script. Read the file as a multiline input command.
	if cmd.Script != "" {
		f, err := os.Open(cmd.Script)
		if err != nil {
			return nil, errors.Wrap(err, "can't open script")
		}

		data, err := io.ReadAll(f)
		if err != nil {
			return nil, errors.Wrap(err, "can't read script")
		}

		task := Task{
			Run: string(data),
			TTY: true,
		}
		if sup.debug {
			task.Run = "set -x;" + task.Run
		}

		if cmd.Stdin {
			task.Input = os.Stdin
		}

		switch {
		case cmd.Once:
			task.Clients = []Client{clients[0]}
			tasks = append(tasks, &task)
		case cmd.Serial > 0:
			// Each "serial" task client group is executed sequentially.
			for i := 0; i < len(clients); i += cmd.Serial {
				j := i + cmd.Serial
				if j > len(clients) {
					j = len(clients)
				}

				taskCopy := task
				taskCopy.Clients = clients[i:j]
				tasks = append(tasks, &taskCopy)
			}
		default:
			task.Clients = clients
			tasks = append(tasks, &task)
		}
	}

	// Local command.
	if cmd.Local != "" {
		local := &LocalhostClient{
			env: env + `export SUP_HOST="localhost";`,
		}

		err := local.Connect("localhost")
		if err != nil {
			return nil, err
		}

		task := &Task{
			Run:     cmd.Local,
			Clients: []Client{local},
			TTY:     true,
		}

		if sup.debug {
			task.Run = "set -x;" + task.Run
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
			TTY: true,
		}
		if sup.debug {
			task.Run = "set -x;" + task.Run
		}

		if cmd.Stdin {
			task.Input = os.Stdin
		}

		switch {
		case cmd.Once:
			task.Clients = []Client{clients[0]}
			tasks = append(tasks, &task)
		case cmd.Serial > 0:
			// Each "serial" task client group is executed sequentially.
			for i := 0; i < len(clients); i += cmd.Serial {
				j := i + cmd.Serial
				if j > len(clients) {
					j = len(clients)
				}

				taskCopy := task
				taskCopy.Clients = clients[i:j]
				tasks = append(tasks, &taskCopy)
			}
		default:
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
