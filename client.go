package sup

type Client interface {
	Connect(host string) error
	Run(task *Task) error
	Wait() error
	Close() error
	Prefix() string
	Write(p []byte) (n int, err error)
	WriteClose() error
}
