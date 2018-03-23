package main

import (
	"testing"
)

func TestSSH(t *testing.T) {
	outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 3)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	input := `
---
version: 0.4

networks:
  local:
    hosts:
    - server0
    - server2

commands:
  test:
    run: echo "Hey over there"
  test2:
    run: echo "Hey again"

targets:
  t:
  - test
  - test2
`
	args := []string{"--sshconfig", sshConfigPath, "local", "t"}
	if err := runSupfile(args, []byte(input)); err != nil {
		t.Fatal(err)
	}

	m := newMatcher(outputs, t)
	m.expectActivityOnServers(0, 2)
	m.expectNoActivityOnServers(1)
	m.expectExportOnActiveServers(`SUP_NETWORK="local"`)
	m.expectCommandOnActiveServers(`echo "Hey over there"`)
}
