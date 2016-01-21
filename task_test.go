package sup

import (
	"strings"
	"testing"
)

func TestCreateTask_Parse_Script(t *testing.T) {
	cmd := &Command{
		Script: "./test/test",
	}
	clients := []Client{&SSHClient{}}
	env := ""

	task, _ := CreateTasks(cmd, clients, env)

	if len(task) < 1 {
		t.Errorf("Failed to parse task")
	}

	// Content is in test/script.sh
	if strings.Trim(task[0].Run, "\n") != "echo 'test'" {
		t.Errorf("Fail to read content of script. Expected: %s Got: %s.", "echo 'test'", task[0].Run)
	}
}

func TestCreateTask_Parse_Script_With_Argument(t *testing.T) {
	cmd := &Command{
		Script: "./test/test -t test",
	}
	clients := []Client{&SSHClient{}}
	env := ""

	task, _ := CreateTasks(cmd, clients, env)

	if len(task) < 1 {
		t.Errorf("Failed to parse task")
	}

	// Content is in test/script.sh
	if strings.Trim(task[0].Run, "\n") != "echo 'test'" {
		t.Errorf("Fail to read content of script. Expected: %s Got: %s.", "echo 'test'", task[0].Run)
	}
}
