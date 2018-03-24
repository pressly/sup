package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"testing"
	"time"
)

var (
	testErrStream = ioutil.Discard
)

func TestInvalidYaml(t *testing.T) {
	t.Parallel()

	input := `
---
version: 0.4
efewf we
we	kp	re
`
	if err := runSupfile(testErrStream, options{}, []string{}, []byte(input)); err == nil {
		t.Fatal("Expected an error")
	}
}

func TestNoNetworkSpecified(t *testing.T) {
	t.Parallel()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server2

commands:
  step1:
    run: echo "Hey over there"
`
	if err := runSupfile(testErrStream, options{}, []string{}, []byte(input)); err != ErrUsage {
		t.Fatal(err)
	}
}

func TestUnknownNetwork(t *testing.T) {
	t.Parallel()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server2

commands:
  step1:
    run: echo "Hey over there"
`
	if err := runSupfile(testErrStream, options{}, []string{"production"}, []byte(input)); err != ErrUnknownNetwork {
		t.Fatal(err)
	}
}

func TestNoHosts(t *testing.T) {
	t.Parallel()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:

commands:
  step1:
    run: echo "Hey over there"
`
	if err := runSupfile(testErrStream, options{}, []string{"staging"}, []byte(input)); err != ErrNetworkNoHosts {
		t.Fatal(err)
	}
}

func TestNoCommand(t *testing.T) {
	t.Parallel()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server2

commands:
  step1:
    run: echo "Hey over there"
`
	if err := runSupfile(testErrStream, options{}, []string{"staging"}, []byte(input)); err != ErrUsage {
		t.Fatal(err)
	}
}

func TestNonexistentCommandOrTarget(t *testing.T) {
	t.Parallel()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server2

commands:
  step1:
    run: echo "Hey over there"
  step2:
    run: echo "Hey again"

targets:
  walk:
  - step1
  - step2
`
	if err := runSupfile(testErrStream, options{}, []string{"staging", "step5"}, []byte(input)); err == nil {
		t.Fatal("Expected an error")
	} else if !strings.Contains(err.Error(), ErrCmd.Error()) {
		t.Fatalf("Expected `%v` error, got `%v`", ErrCmd, err)
	}
}

func TestTargetReferencingNonexistentCommand(t *testing.T) {
	t.Parallel()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server2

commands:
  step1:
    run: echo "Hey over there"
  step2:
    run: echo "Hey again"

targets:
  walk:
  - step1
  - step2
  - step3
`
	if err := runSupfile(testErrStream, options{}, []string{"staging", "walk"}, []byte(input)); err == nil {
		t.Fatal("Expected an error")
	} else if !strings.Contains(err.Error(), ErrCmd.Error()) {
		t.Fatalf("Expected `%v` error, got `%v`", ErrCmd, err)
	}
}

func TestOneCommand(t *testing.T) {
	t.Parallel()

	outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 3)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server2

commands:
  step1:
    run: echo "Hey over there"
`
	options := options{
		sshConfig: sshConfigPath,
	}
	args := []string{"staging", "step1"}
	if err := runSupfile(testErrStream, options, args, []byte(input)); err != nil {
		t.Fatal(err)
	}

	m := newMatcher(outputs, t)
	m.expectActivityOnServers(0, 2)
	m.expectNoActivityOnServers(1)
	m.expectCommandOnActiveServers(`echo "Hey over there"`)
}

func TestSequenceOfCommands(t *testing.T) {
	t.Parallel()

	outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 3)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server2

commands:
  step1:
    run: echo "Hey over there"
  step2:
    run: echo "Hey again"
`
	options := options{
		sshConfig: sshConfigPath,
	}
	args := []string{"staging", "step1", "step2"}
	if err := runSupfile(testErrStream, options, args, []byte(input)); err != nil {
		t.Fatal(err)
	}

	m := newMatcher(outputs, t)
	m.expectActivityOnServers(0, 2)
	m.expectNoActivityOnServers(1)
	m.expectCommandOnActiveServers(`echo "Hey over there"`)
	m.expectCommandOnActiveServers(`echo "Hey again"`)
}

func TestTarget(t *testing.T) {
	t.Parallel()

	outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 3)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server2

commands:
  step1:
    run: echo "Hey over there"
  step2:
    run: echo "Hey again"

targets:
  walk:
  - step1
  - step2
`
	options := options{
		sshConfig: sshConfigPath,
	}
	args := []string{"staging", "walk"}
	if err := runSupfile(testErrStream, options, args, []byte(input)); err != nil {
		t.Fatal(err)
	}

	m := newMatcher(outputs, t)
	m.expectActivityOnServers(0, 2)
	m.expectNoActivityOnServers(1)
	m.expectCommandOnActiveServers(`echo "Hey over there"`)
	m.expectCommandOnActiveServers(`echo "Hey again"`)
}

func TestOnlyHosts(t *testing.T) {
	t.Parallel()

	outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 3)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server1
    - server2

commands:
  step1:
    run: echo "Hey over there"
`
	options := options{
		sshConfig: sshConfigPath,
		onlyHosts: "server2",
	}
	args := []string{"staging", "step1"}
	if err := runSupfile(testErrStream, options, args, []byte(input)); err != nil {
		t.Fatal(err)
	}

	m := newMatcher(outputs, t)
	m.expectActivityOnServers(2)
	m.expectNoActivityOnServers(0, 1)
	m.expectCommandOnActiveServers(`echo "Hey over there"`)
}

func TestOnlyHostsEmpty(t *testing.T) {
	t.Parallel()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server1
    - server2

commands:
  step1:
    run: echo "Hey over there"
`
	options := options{
		onlyHosts: "server42",
	}
	if err := runSupfile(testErrStream, options, []string{"staging", "step1"}, []byte(input)); err == nil {
		t.Fatal("Expected an error")
	} else if !strings.Contains(err.Error(), "no hosts match") {
		t.Fatalf("Expected a different error, got `%v`", err)
	}
}

func TestOnlyHostsInvalid(t *testing.T) {
	t.Parallel()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server1
    - server2

commands:
  step1:
    run: echo "Hey over there"
`
	options := options{
		onlyHosts: "server(",
	}
	if err := runSupfile(testErrStream, options, []string{"staging", "step1"}, []byte(input)); err == nil {
		t.Fatal("Expected an error")
	} else if !strings.Contains(err.Error(), "error parsing regexp") {
		t.Fatalf("Expected a different error, got `%v`", err)
	}
}

func TestExceptHosts(t *testing.T) {
	t.Parallel()

	outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 3)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server1
    - server2

commands:
  step1:
    run: echo "Hey over there"
`
	options := options{
		sshConfig:   sshConfigPath,
		exceptHosts: "server(1|2)",
	}
	args := []string{"staging", "step1"}
	if err := runSupfile(testErrStream, options, args, []byte(input)); err != nil {
		t.Fatal(err)
	}

	m := newMatcher(outputs, t)
	m.expectActivityOnServers(0)
	m.expectNoActivityOnServers(1, 2)
	m.expectCommandOnActiveServers(`echo "Hey over there"`)
}

func TestExceptHostsEmpty(t *testing.T) {
	t.Parallel()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server1
    - server2

commands:
  step1:
    run: echo "Hey over there"
`
	options := options{
		exceptHosts: "server",
	}
	if err := runSupfile(testErrStream, options, []string{"staging", "step1"}, []byte(input)); err == nil {
		t.Fatal("Expected an error")
	} else if !strings.Contains(err.Error(), "no hosts left") {
		t.Fatalf("Expected a different error, got `%v`", err)
	}
}

func TestExceptHostsInvalid(t *testing.T) {
	t.Parallel()

	input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server1
    - server2

commands:
  step1:
    run: echo "Hey over there"
`
	options := options{
		exceptHosts: "server(",
	}
	if err := runSupfile(testErrStream, options, []string{"staging", "step1"}, []byte(input)); err == nil {
		t.Fatal("Expected an error")
	} else if !strings.Contains(err.Error(), "error parsing regexp") {
		t.Fatalf("Expected a different error, got `%v`", err)
	}
}

func TestInventory(t *testing.T) {
	t.Parallel()

	outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 3)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	input := `
---
version: 0.4

networks:
  staging:
    inventory: array=( 0 2 ); for i in "${array[@]}"; do printf "server$i\n\n# comment\n"; done

commands:
  step1:
    run: echo "Hey over there"
`
	options := options{
		sshConfig: sshConfigPath,
	}
	args := []string{"staging", "step1"}
	if err := runSupfile(testErrStream, options, args, []byte(input)); err != nil {
		t.Fatal(err)
	}

	m := newMatcher(outputs, t)
	m.expectActivityOnServers(0, 2)
	m.expectNoActivityOnServers(1)
	m.expectCommandOnActiveServers(`echo "Hey over there"`)
}

func TestFailedInventory(t *testing.T) {
	t.Parallel()

	input := `
---
version: 0.4

networks:
  staging:
    inventory: this won't compile

commands:
  step1:
    run: echo "Hey over there"
`
	args := []string{"staging", "step1"}
	if err := runSupfile(testErrStream, options{}, args, []byte(input)); err == nil {
		t.Fatal("Expected an error")
	}
}

func TestSupVariables(t *testing.T) {
	t.Parallel()

	// these tests need to run in order because they mess with env vars
	t.Run("default", func(t *testing.T) {
		if time.Now().Hour() == 23 && time.Now().Minute() == 59 {
			t.Skip("Skipping test")
		}

		outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 2)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()

		input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server1

commands:
  step1:
    run: echo "Hey over there"
`
		options := options{
			sshConfig: sshConfigPath,
		}
		if err := runSupfile(testErrStream, options, []string{"staging", "step1"}, []byte(input)); err != nil {
			t.Fatal(err)
		}
		currentUser, err := user.Current()
		if err != nil {
			t.Fatal(err)
		}
		m := newMatcher(outputs, t)
		m.expectActivityOnServers(0, 1)
		m.expectExportOnActiveServers(`SUP_NETWORK="staging"`)
		m.expectExportOnActiveServers(`SUP_ENV=""`)
		m.expectExportOnActiveServers(fmt.Sprintf(`SUP_USER="%s"`, currentUser.Name))
		m.expectExportRegexpOnActiveServers(`SUP_HOST="localhost:\d+"`)
	})

	t.Run("default SUP_TIME", func(t *testing.T) {

		if time.Now().Hour() == 23 && time.Now().Minute() == 59 {
			t.Skip("Skipping SUP_TIME test because it might fail around midnight")
		}

		outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 2)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()

		input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server1

commands:
  step1:
    run: echo "Hey over there"
`
		options := options{
			sshConfig: sshConfigPath,
		}
		if err := runSupfile(testErrStream, options, []string{"staging", "step1"}, []byte(input)); err != nil {
			t.Fatal(err)
		}
		m := newMatcher(outputs, t)
		m.expectActivityOnServers(0, 1)
		m.expectExportRegexpOnActiveServers(
			fmt.Sprintf(
				`SUP_TIME="%4d-%02d-%02dT[0-2][0-9]:[0-5][0-9]:[0-5][0-9]Z"`,
				time.Now().Year(),
				time.Now().Month(),
				time.Now().Day(),
			),
		)
	})

	t.Run("SUP_TIME env var set", func(t *testing.T) {
		outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 2)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()

		input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server1

commands:
  step1:
    run: echo "Hey over there"
`
		os.Setenv("SUP_TIME", "now")
		options := options{
			sshConfig: sshConfigPath,
		}
		if err := runSupfile(testErrStream, options, []string{"staging", "step1"}, []byte(input)); err != nil {
			t.Fatal(err)
		}
		m := newMatcher(outputs, t)
		m.expectActivityOnServers(0, 1)
		m.expectExportOnActiveServers(`SUP_TIME="now"`)
	})

	t.Run("SUP_USER env var set", func(t *testing.T) {
		outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 2)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()

		input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0
    - server1

commands:
  step1:
    run: echo "Hey over there"
`
		os.Setenv("SUP_USER", "sup_rules")
		options := options{
			sshConfig: sshConfigPath,
		}
		if err := runSupfile(testErrStream, options, []string{"staging", "step1"}, []byte(input)); err != nil {
			t.Fatal(err)
		}
		m := newMatcher(outputs, t)
		m.expectActivityOnServers(0, 1)
		m.expectExportOnActiveServers(`SUP_USER="sup_rules"`)
	})
}

func TestInvalidVars(t *testing.T) {
	t.Parallel()

	_, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 2)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	input := `
---
version: 0.4

env:
  TODAYS_SPECIAL: this won't compile

networks:
  staging:
    hosts:
    - server0
    - server1

commands:
  step1:
    run: echo "Hey over there"
`
	options := options{
		sshConfig: sshConfigPath,
	}
	args := []string{"staging", "step1"}
	if err := runSupfile(testErrStream, options, args, []byte(input)); err == nil {
		t.Fatal("Expected an error")
	}

}

func TestEnvLevelVars(t *testing.T) {
	t.Parallel()

	outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 2)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	input := `
---
version: 0.4

env:
  TODAYS_SPECIAL: "dog milk"

networks:
  staging:
    hosts:
    - server0
    - server1

commands:
  step1:
    run: echo "Hey over there"
`
	options := options{
		sshConfig: sshConfigPath,
	}
	args := []string{"staging", "step1"}
	if err := runSupfile(testErrStream, options, args, []byte(input)); err != nil {
		t.Fatal(err)
	}

	m := newMatcher(outputs, t)
	m.expectActivityOnServers(0, 1)
	m.expectExportOnActiveServers(`TODAYS_SPECIAL="dog milk"`)
	m.expectCommandOnActiveServers(`echo "Hey over there"`)
}

func TestNetworkLevelVars(t *testing.T) {
	t.Parallel()

	outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 2)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	input := `
---
version: 0.4

env:
  TODAYS_SPECIAL: "dog milk"

networks:
  staging:
    env:
      TODAYS_SPECIAL: "Trout a la Crème"
    hosts:
    - server0
    - server1

commands:
  step1:
    run: echo "Hey over there"
`
	options := options{
		sshConfig: sshConfigPath,
	}
	args := []string{"staging", "step1"}
	if err := runSupfile(testErrStream, options, args, []byte(input)); err != nil {
		t.Fatal(err)
	}

	m := newMatcher(outputs, t)
	m.expectActivityOnServers(0, 1)
	m.expectExportOnActiveServers(`TODAYS_SPECIAL="Trout a la Crème"`)
	m.expectCommandOnActiveServers(`echo "Hey over there"`)
}

func TestCommandLineLevelVars(t *testing.T) {
	t.Parallel()

	outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 2)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	input := `
---
version: 0.4

env:
  TODAYS_SPECIAL: "dog milk"

networks:
  staging:
    env:
      TODAYS_SPECIAL: "Trout a la Crème"
    hosts:
    - server0
    - server1

commands:
  step1:
    run: echo "Hey over there"
`
	options := options{
		sshConfig: sshConfigPath,
		envVars:   []string{"IM_HERE", "TODAYS_SPECIAL=Gazpacho"},
	}
	args := []string{"staging", "step1"}
	if err := runSupfile(testErrStream, options, args, []byte(input)); err != nil {
		t.Fatal(err)
	}

	m := newMatcher(outputs, t)
	m.expectActivityOnServers(0, 1)
	m.expectExportOnActiveServers(`IM_HERE=""`)
	m.expectExportOnActiveServers(`TODAYS_SPECIAL="Gazpacho"`)
	m.expectExportOnActiveServers(`SUP_ENV="-e TODAYS_SPECIAL="Gazpacho""`)
	m.expectCommandOnActiveServers(`echo "Hey over there"`)
}

func TestEmptyCommandLineLevelVars(t *testing.T) {
	t.Parallel()

	outputs, sshConfigPath, cleanup, err := setupMockEnv("ssh_config", 2)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	input := `
---
version: 0.4

env:
  TODAYS_SPECIAL: "dog milk"

networks:
  staging:
    env:
      TODAYS_SPECIAL: "Trout a la Crème"
    hosts:
    - server0
    - server1

commands:
  step1:
    run: echo "Hey over there"
`
	options := options{
		sshConfig: sshConfigPath,
		envVars:   []string{""},
	}
	args := []string{"staging", "step1"}
	if err := runSupfile(testErrStream, options, args, []byte(input)); err != nil {
		t.Fatal(err)
	}

	m := newMatcher(outputs, t)
	m.expectActivityOnServers(0, 1)
	m.expectExportOnActiveServers(`TODAYS_SPECIAL="Trout a la Crème"`)
	m.expectExportOnActiveServers(`SUP_ENV=""`)
	m.expectCommandOnActiveServers(`echo "Hey over there"`)
}
