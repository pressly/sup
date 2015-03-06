package stackup

type Client interface {
	Connect(host string) error
	Run(task Task) error
	Wait() error
	Close() error
	Prefix() string
}
