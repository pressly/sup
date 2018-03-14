package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// matcher defines high level expectations over a collection of output buffers
type matcher struct {
	outputs       []bytes.Buffer
	t             *testing.T
	activeServers []int
}

func newMatcher(outputs []bytes.Buffer, t *testing.T) matcher {
	return matcher{
		outputs: outputs,
		t:       t,
	}
}

func (m *matcher) expectActivityOnServers(servers ...int) {
	m.activeServers = servers
	m.onEachActiveServer(func(server int, output string) {
		if len(output) == 0 {
			m.t.Errorf("expected activity on server #%d", server)
		}
	})
}
func (m *matcher) expectNoActivityOnServers(servers ...int) {
	for _, server := range servers {
		if server >= len(m.outputs) || server < 0 {
			m.t.Errorf("output from server #%d not provided", server)
			return
		}
		output := m.outputs[server]
		if output.Len() > 0 {
			m.t.Errorf("expected no activity on server #%d:\n%s", server, output.String())
		}
	}
}

func (m matcher) expectExportOnActiveServers(export string) {
	m.onEachActiveServer(func(server int, output string) {
		for i, executed := range strings.Split(output, "\n") {
			if !strings.Contains(executed, fmt.Sprintf("export %s;", export)) {
				m.t.Errorf(
					"command #%d on server #%d does not export `%s`:\n%s",
					i,
					server,
					export,
					executed,
				)
			}
		}
	})
}

func (m matcher) expectCommandOnActiveServers(command string) {
	m.onEachActiveServer(func(server int, output string) {
		for _, executed := range strings.Split(output, "\n") {
			if strings.HasSuffix(executed, fmt.Sprintf(";%s", command)) {
				return
			}
		}
		m.t.Errorf("no command on server #%d executed `%s`", server, command)
	})
}

func (m matcher) onEachActiveServer(expectation func(server int, output string)) {
	for _, server := range m.activeServers {
		if server >= len(m.outputs) || server < 0 {
			m.t.Errorf("output from server #%d not provided", server)
			return
		}

		output := m.outputs[server]
		expectation(server, strings.TrimSpace(output.String()))
	}
}
