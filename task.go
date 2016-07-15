package sup

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// Task represents a set of commands to be run.
type Task struct {
	Interpreter string
	Run         string
	Input       io.Reader
	Clients     []Client
}

func CreateTasks(cmd *Command, clients []Client, env string) ([]*Task, error) {
	var tasks []*Task

	// Anything to upload?
	for _, upload := range cmd.Upload {
		task := Task{
			Run:   RemoteTarCommand(upload.Dst),
			Input: NewTarStreamReader(upload.Src, upload.Exc, env),
		}

		if cmd.Once {
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

	// Script. Read the file as a multiline input command.
	if cmd.Script != "" {
		f, err := os.Open(cmd.Script)
		if err != nil {
			return nil, err
		}
		rd := bufio.NewReader(f)
		fLine, _, err := rd.ReadLine()
		if err != nil {
			return nil, err
		}
		var hashbang, data []byte
		if len(fLine) > 2 && fLine[0] == '#' && fLine[1] == '!' {
			hashbang = fLine[2:len(fLine)]
		} else {
			rd.Reset(f)
		}
		for {
			line, _, err := rd.ReadLine()
			if len(line) > 0 {
				data = append(data, line...)
				data = append(data, '\n')
			}

			if err != nil {
				if err != io.EOF {
					return nil, err
				}
				break
			}
		}

		task := Task{
			Interpreter: string(hashbang),
			Run:         string(data),
		}
		if cmd.Stdin {
			task.Input = os.Stdin
		}
		if cmd.Once {
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
		if cmd.Once {
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
