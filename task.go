package sup

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
)

type Task struct {
	Run     string
	Input   io.Reader
	Clients []Client
	TTY     bool
}

type finishReader struct {
	io.Reader
	bar *pb.ProgressBar
}

func (r *finishReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if err == io.EOF {
		r.bar.Finish()
	}
	return
}

func (sup *Stackup) createTasks(cmd *Command, clients []Client, env string) ([]*Task, error) {
	var tasks []*Task

	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "resolving CWD failed")
	}

	for _, upload := range cmd.Upload {
		uploadFile, err := ResolveLocalPath(cwd, upload.Src, env)
		if err != nil {
			return nil, errors.Wrap(err, "upload src: "+upload.Src)
		}

		uploadDst, err := ResolveLocalPath(cwd, upload.Dst, env)
		if err != nil {
			return nil, errors.Wrap(err, "upload dst: "+upload.Dst)
		}

		tarBaseDir := "."
		tarTarget := uploadFile

		if upload.Flatten {
			tarBaseDir = filepath.Dir(uploadFile)
			tarTarget = filepath.Base(uploadFile)
		}

		uploadTarReader, err := NewTarStreamReader(cwd, tarBaseDir, tarTarget, upload.Exc)
		if err != nil {
			return nil, errors.Wrap(err, "upload: "+upload.Src)
		}

		bar := pb.New64(0)
		bar.Set(pb.Bytes, true)
		bar.Set("prefix", fmt.Sprintf("Uploading %s: ", uploadFile))
		bar.SetTemplateString(`{{string . "prefix"}}{{counters . }} {{speed . }}`)
		bar.Start()

		proxyReader := &finishReader{
			Reader: bar.NewProxyReader(uploadTarReader),
			bar:    bar,
		}

		task := Task{
			Run:   RemoteTarCommand(uploadDst, upload.Sudo),
			Input: proxyReader,
			TTY:   false,
		}

		if cmd.Once {
			task.Clients = []Client{clients[0]}
			tasks = append(tasks, &task)
		} else if cmd.Serial > 0 {
			for i := 0; i < len(clients); i += cmd.Serial {
				j := min(i+cmd.Serial, len(clients))
				copy := task
				copy.Clients = clients[i:j]
				tasks = append(tasks, &copy)
			}
		} else {
			task.Clients = clients
			tasks = append(tasks, &task)
		}
	}

	if cmd.Script != "" {
		f, err := os.Open(cmd.Script)
		if err != nil {
			return nil, errors.Wrap(err, "can't open script")
		}
		data, err := ioutil.ReadAll(f)
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
		if cmd.Once {
			task.Clients = []Client{clients[0]}
			tasks = append(tasks, &task)
		} else if cmd.Serial > 0 {
			for i := 0; i < len(clients); i += cmd.Serial {
				j := min(i+cmd.Serial, len(clients))
				copy := task
				copy.Clients = clients[i:j]
				tasks = append(tasks, &copy)
			}
		} else {
			task.Clients = clients
			tasks = append(tasks, &task)
		}
	}

	if cmd.Local != "" {
		local := &LocalhostClient{
			env: env + `export SUP_HOST="localhost";`,
		}
		local.Connect("localhost")
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
		if cmd.Once {
			task.Clients = []Client{clients[0]}
			tasks = append(tasks, &task)
		} else if cmd.Serial > 0 {
			for i := 0; i < len(clients); i += cmd.Serial {
				j := min(i+cmd.Serial, len(clients))
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
