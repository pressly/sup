package client

type HybridClient struct {
	Remote    SSHClient
	Localhost LocalhostClient
}

func (c *HybridClient) Connect(host string) error {
	return nil
}

func (c *HybridClient) Run(task Task) error {
	return nil
}

func (c *HybridClient) Wait() error {
	return nil
}

func (c *HybridClient) Close() error {
	return nil
}

func (c *HybridClient) Prefix() string {
	return ""
}
