package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adammck/venv"
)

const (
	envTestUser = "sup_test_user"
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

	withTmpDir(t, input, func(dirname string) {
		options := options{
			dirname: dirname,
			env:     testEnv(),
		}
		if err := runSupfile(testErrStream, options, []string{}); err == nil {
			t.Fatal("Expected an error")
		}
	})
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
	withTmpDir(t, input, func(dirname string) {
		options := options{
			dirname: dirname,
			env:     testEnv(),
		}
		if err := runSupfile(testErrStream, options, []string{}); err != ErrUsage {
			t.Fatal(err)
		}
	})
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
	withTmpDir(t, input, func(dirname string) {
		options := options{
			dirname: dirname,
		}
		if err := runSupfile(testErrStream, options, []string{"production"}); err != ErrUnknownNetwork {
			t.Fatal(err)
		}
	})
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
	withTmpDir(t, input, func(dirname string) {
		options := options{
			dirname: dirname,
			env:     testEnv(),
		}
		if err := runSupfile(testErrStream, options, []string{"staging"}); err != ErrNetworkNoHosts {
			t.Fatal(err)
		}
	})
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
	withTmpDir(t, input, func(dirname string) {
		options := options{
			dirname: dirname,
			env:     testEnv(),
		}
		if err := runSupfile(testErrStream, options, []string{"staging"}); err != ErrUsage {
			t.Fatal(err)
		}
	})
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
	withTmpDir(t, input, func(dirname string) {
		options := options{
			dirname: dirname,
			env:     testEnv(),
		}
		if err := runSupfile(testErrStream, options, []string{"staging", "step5"}); err == nil {
			t.Fatal("Expected an error")
		} else if !strings.Contains(err.Error(), ErrCmd.Error()) {
			t.Fatalf("Expected `%v` error, got `%v`", ErrCmd, err)
		}
	})
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
	withTmpDir(t, input, func(dirname string) {
		options := options{
			dirname: dirname,
			env:     testEnv(),
		}
		if err := runSupfile(testErrStream, options, []string{"staging", "walk"}); err == nil {
			t.Fatal("Expected an error")
		} else if !strings.Contains(err.Error(), ErrCmd.Error()) {
			t.Fatalf("Expected `%v` error, got `%v`", ErrCmd, err)
		}
	})
}

func TestOneCommand(t *testing.T) {
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
	withTmpDir(t, input, func(dirname string) {
		outputs, options, err := setupMockEnv(dirname, 3)
		if err != nil {
			t.Fatal(err)
		}

		args := []string{"staging", "step1"}
		if err := runSupfile(testErrStream, options, args); err != nil {
			t.Fatal(err)
		}

		m := newMatcher(outputs, t)
		m.expectActivityOnServers(0, 2)
		m.expectNoActivityOnServers(1)
		m.expectCommandOnActiveServers(`echo "Hey over there"`)
	})
}

func TestSequenceOfCommands(t *testing.T) {
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
`
	withTmpDir(t, input, func(dirname string) {
		outputs, options, err := setupMockEnv(dirname, 3)
		if err != nil {
			t.Fatal(err)
		}

		args := []string{"staging", "step1", "step2"}
		if err := runSupfile(testErrStream, options, args); err != nil {
			t.Fatal(err)
		}

		m := newMatcher(outputs, t)
		m.expectActivityOnServers(0, 2)
		m.expectNoActivityOnServers(1)
		m.expectCommandOnActiveServers(`echo "Hey over there"`)
		m.expectCommandOnActiveServers(`echo "Hey again"`)
	})
}

func TestTarget(t *testing.T) {
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
	withTmpDir(t, input, func(dirname string) {
		outputs, options, err := setupMockEnv(dirname, 3)
		if err != nil {
			t.Fatal(err)
		}

		args := []string{"staging", "walk"}
		if err := runSupfile(testErrStream, options, args); err != nil {
			t.Fatal(err)
		}

		m := newMatcher(outputs, t)
		m.expectActivityOnServers(0, 2)
		m.expectNoActivityOnServers(1)
		m.expectCommandOnActiveServers(`echo "Hey over there"`)
		m.expectCommandOnActiveServers(`echo "Hey again"`)
	})
}

func TestOnlyHosts(t *testing.T) {
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
	withTmpDir(t, input, func(dirname string) {
		outputs, options, err := setupMockEnv(dirname, 3)
		if err != nil {
			t.Fatal(err)
		}

		options.onlyHosts = "server2"
		args := []string{"staging", "step1"}
		if err := runSupfile(testErrStream, options, args); err != nil {
			t.Fatal(err)
		}

		m := newMatcher(outputs, t)
		m.expectActivityOnServers(2)
		m.expectNoActivityOnServers(0, 1)
		m.expectCommandOnActiveServers(`echo "Hey over there"`)
	})
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
	withTmpDir(t, input, func(dirname string) {
		options := options{
			dirname:   dirname,
			onlyHosts: "server42",
			env:       venv.Mock(),
		}
		if err := runSupfile(testErrStream, options, []string{"staging", "step1"}); err == nil {
			t.Fatal("Expected an error")
		} else if !strings.Contains(err.Error(), "no hosts match") {
			t.Fatalf("Expected a different error, got `%v`", err)
		}
	})
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
	withTmpDir(t, input, func(dirname string) {
		options := options{
			dirname:   dirname,
			onlyHosts: "server(",
			env:       venv.Mock(),
		}
		if err := runSupfile(testErrStream, options, []string{"staging", "step1"}); err == nil {
			t.Fatal("Expected an error")
		} else if !strings.Contains(err.Error(), "error parsing regexp") {
			t.Fatalf("Expected a different error, got `%v`", err)
		}
	})
}

func TestExceptHosts(t *testing.T) {
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
	withTmpDir(t, input, func(dirname string) {
		outputs, options, err := setupMockEnv(dirname, 3)
		if err != nil {
			t.Fatal(err)
		}
		options.exceptHosts = "server(1|2)"
		args := []string{"staging", "step1"}
		if err := runSupfile(testErrStream, options, args); err != nil {
			t.Fatal(err)
		}

		m := newMatcher(outputs, t)
		m.expectActivityOnServers(0)
		m.expectNoActivityOnServers(1, 2)
		m.expectCommandOnActiveServers(`echo "Hey over there"`)
	})
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
	withTmpDir(t, input, func(dirname string) {
		options := options{
			dirname:     dirname,
			exceptHosts: "server",
			env:         venv.Mock(),
		}
		if err := runSupfile(testErrStream, options, []string{"staging", "step1"}); err == nil {
			t.Fatal("Expected an error")
		} else if !strings.Contains(err.Error(), "no hosts left") {
			t.Fatalf("Expected a different error, got `%v`", err)
		}
	})
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
	withTmpDir(t, input, func(dirname string) {
		options := options{
			dirname:     dirname,
			exceptHosts: "server(",
			env:         venv.Mock(),
		}
		if err := runSupfile(testErrStream, options, []string{"staging", "step1"}); err == nil {
			t.Fatal("Expected an error")
		} else if !strings.Contains(err.Error(), "error parsing regexp") {
			t.Fatalf("Expected a different error, got `%v`", err)
		}
	})
}

func TestInventory(t *testing.T) {
	t.Parallel()

	input := `
---
version: 0.4

networks:
  staging:
    inventory: printf "server0\n# comment\n\nserver2\n"

commands:
  step1:
    run: echo "Hey over there"
`
	withTmpDir(t, input, func(dirname string) {
		outputs, options, err := setupMockEnv(dirname, 3)
		if err != nil {
			t.Fatal(err)
		}

		args := []string{"staging", "step1"}
		if err := runSupfile(testErrStream, options, args); err != nil {
			t.Fatal(err)
		}

		m := newMatcher(outputs, t)
		m.expectActivityOnServers(0, 2)
		m.expectNoActivityOnServers(1)
		m.expectCommandOnActiveServers(`echo "Hey over there"`)
	})
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
	withTmpDir(t, input, func(dirname string) {
		options := options{
			dirname: dirname,
			env:     testEnv(),
		}
		args := []string{"staging", "step1"}
		if err := runSupfile(testErrStream, options, args); err == nil {
			t.Fatal("Expected an error")
		}
	})
}

func TestSupVariables(t *testing.T) {
	t.Parallel()

	// these tests need to run in order because they mess with env vars
	t.Run("default", func(t *testing.T) {
		if time.Now().Hour() == 23 && time.Now().Minute() == 59 {
			t.Skip("Skipping test")
		}

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
		withTmpDir(t, input, func(dirname string) {
			outputs, options, err := setupMockEnv(dirname, 2)
			if err != nil {
				t.Fatal(err)
			}

			if err := runSupfile(testErrStream, options, []string{"staging", "step1"}); err != nil {
				t.Fatal(err)
			}
			m := newMatcher(outputs, t)
			m.expectActivityOnServers(0, 1)
			m.expectExportOnActiveServers(`SUP_NETWORK="staging"`)
			m.expectExportOnActiveServers(`SUP_ENV=""`)
			m.expectExportOnActiveServers(fmt.Sprintf(`SUP_USER="%s"`, envTestUser))
			m.expectExportRegexpOnActiveServers(`SUP_HOST="localhost:\d+"`)
		})
	})

	t.Run("default SUP_TIME", func(t *testing.T) {

		if time.Now().Hour() == 23 && time.Now().Minute() == 59 {
			t.Skip("Skipping SUP_TIME test because it might fail around midnight")
		}

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
		withTmpDir(t, input, func(dirname string) {
			outputs, options, err := setupMockEnv(dirname, 2)
			if err != nil {
				t.Fatal(err)
			}

			if err := runSupfile(testErrStream, options, []string{"staging", "step1"}); err != nil {
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
	})

	t.Run("SUP_TIME env var set", func(t *testing.T) {

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
		withTmpDir(t, input, func(dirname string) {
			outputs, options, err := setupMockEnv(dirname, 2)
			if err != nil {
				t.Fatal(err)
			}
			options.env.Setenv("SUP_TIME", "now")

			if err := runSupfile(testErrStream, options, []string{"staging", "step1"}); err != nil {
				t.Fatal(err)
			}
			m := newMatcher(outputs, t)
			m.expectActivityOnServers(0, 1)
			m.expectExportOnActiveServers(`SUP_TIME="now"`)
		})
	})

	t.Run("SUP_USER env var set", func(t *testing.T) {

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
		withTmpDir(t, input, func(dirname string) {
			outputs, options, err := setupMockEnv(dirname, 2)
			if err != nil {
				t.Fatal(err)
			}
			options.env.Setenv("SUP_USER", "sup_rules")

			if err := runSupfile(testErrStream, options, []string{"staging", "step1"}); err != nil {
				t.Fatal(err)
			}
			m := newMatcher(outputs, t)
			m.expectActivityOnServers(0, 1)
			m.expectExportOnActiveServers(`SUP_USER="sup_rules"`)
		})
	})
}

func TestInvalidVars(t *testing.T) {
	t.Parallel()

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
	withTmpDir(t, input, func(dirname string) {
		_, options, err := setupMockEnv(dirname, 2)
		if err != nil {
			t.Fatal(err)
		}

		args := []string{"staging", "step1"}
		if err := runSupfile(testErrStream, options, args); err == nil {
			t.Fatal("Expected an error")
		}

	})
}

func TestEnvLevelVars(t *testing.T) {
	t.Parallel()

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
	withTmpDir(t, input, func(dirname string) {
		outputs, options, err := setupMockEnv(dirname, 2)
		if err != nil {
			t.Fatal(err)
		}

		args := []string{"staging", "step1"}
		if err := runSupfile(testErrStream, options, args); err != nil {
			t.Fatal(err)
		}

		m := newMatcher(outputs, t)
		m.expectActivityOnServers(0, 1)
		m.expectExportOnActiveServers(`TODAYS_SPECIAL="dog milk"`)
		m.expectCommandOnActiveServers(`echo "Hey over there"`)
	})
}

func TestNetworkLevelVars(t *testing.T) {
	t.Parallel()

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
	withTmpDir(t, input, func(dirname string) {
		outputs, options, err := setupMockEnv(dirname, 2)
		if err != nil {
			t.Fatal(err)
		}

		args := []string{"staging", "step1"}
		if err := runSupfile(testErrStream, options, args); err != nil {
			t.Fatal(err)
		}

		m := newMatcher(outputs, t)
		m.expectActivityOnServers(0, 1)
		m.expectExportOnActiveServers(`TODAYS_SPECIAL="Trout a la Crème"`)
		m.expectCommandOnActiveServers(`echo "Hey over there"`)
	})
}

func TestCommandLineLevelVars(t *testing.T) {
	t.Parallel()

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
	withTmpDir(t, input, func(dirname string) {
		outputs, options, err := setupMockEnv(dirname, 2)
		if err != nil {
			t.Fatal(err)
		}
		options.envVars = []string{"IM_HERE", "TODAYS_SPECIAL=Gazpacho"}
		args := []string{"staging", "step1"}
		if err := runSupfile(testErrStream, options, args); err != nil {
			t.Fatal(err)
		}

		m := newMatcher(outputs, t)
		m.expectActivityOnServers(0, 1)
		m.expectExportOnActiveServers(`IM_HERE=""`)
		m.expectExportOnActiveServers(`TODAYS_SPECIAL="Gazpacho"`)
		m.expectExportOnActiveServers(`SUP_ENV="-e TODAYS_SPECIAL="Gazpacho""`)
		m.expectCommandOnActiveServers(`echo "Hey over there"`)
	})
}

func TestEmptyCommandLineLevelVars(t *testing.T) {
	t.Parallel()

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
	withTmpDir(t, input, func(dirname string) {
		outputs, options, err := setupMockEnv(dirname, 2)
		if err != nil {
			t.Fatal(err)
		}
		options.envVars = []string{""}
		args := []string{"staging", "step1"}
		if err := runSupfile(testErrStream, options, args); err != nil {
			t.Fatal(err)
		}

		m := newMatcher(outputs, t)
		m.expectActivityOnServers(0, 1)
		m.expectExportOnActiveServers(`TODAYS_SPECIAL="Trout a la Crème"`)
		m.expectExportOnActiveServers(`SUP_ENV=""`)
		m.expectCommandOnActiveServers(`echo "Hey over there"`)
	})
}

func TestFileOption(t *testing.T) {
	t.Parallel()

	t.Run("fallbacks to Supfile.yml", func(t *testing.T) {
		t.Parallel()

		input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0

commands:
  step1:
    run: echo "Hey over there"
`
		withTmpDirWithoutSupfile(t, func(dirname string) {
			writeSupfileAs(dirname, "Supfile.yml", input)

			outputs, options, err := setupMockEnv(dirname, 1)
			if err != nil {
				t.Fatal(err)
			}

			args := []string{"staging", "step1"}
			if err := runSupfile(testErrStream, options, args); err != nil {
				t.Fatal(err)
			}

			m := newMatcher(outputs, t)
			m.expectActivityOnServers(0)
			m.expectCommandOnActiveServers(`echo "Hey over there"`)
		})
	})

	t.Run("prefers Supfile over Supfile.yml when not specified", func(t *testing.T) {
		t.Parallel()

		input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0

commands:
  step1:
    run: echo "Default Supfile"
`
		withTmpDir(t, input, func(dirname string) {
			input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0

commands:
  step1:
    run: echo "Supfile.yml"
`
			writeSupfileAs(dirname, "Supfile.yml", input)

			outputs, options, err := setupMockEnv(dirname, 1)
			if err != nil {
				t.Fatal(err)
			}

			args := []string{"staging", "step1"}
			if err := runSupfile(testErrStream, options, args); err != nil {
				t.Fatal(err)
			}

			m := newMatcher(outputs, t)
			m.expectActivityOnServers(0)
			m.expectCommandOnActiveServers(`echo "Default Supfile"`)
		})
	})

	t.Run("can specify arbitrary file", func(t *testing.T) {
		t.Parallel()

		input := `
---
version: 0.4

networks:
  staging:
    hosts:
    - server0

commands:
  step1:
    run: echo "Hey over there"
`
		withTmpDirWithoutSupfile(t, func(dirname string) {
			writeSupfileAs(dirname, "different_file_name", input)

			outputs, options, err := setupMockEnv(dirname, 1)
			if err != nil {
				t.Fatal(err)
			}

			options.supfile = "different_file_name"
			args := []string{"staging", "step1"}
			if err := runSupfile(testErrStream, options, args); err != nil {
				t.Fatal(err)
			}

			m := newMatcher(outputs, t)
			m.expectActivityOnServers(0)
			m.expectCommandOnActiveServers(`echo "Hey over there"`)
		})
	})

	t.Run("fails without a Supfile", func(t *testing.T) {
		t.Parallel()

		withTmpDirWithoutSupfile(t, func(dirname string) {
			options := options{
				dirname: dirname,
			}
			args := []string{"staging", "step1"}
			if err := runSupfile(testErrStream, options, args); err == nil {
				t.Fatal("Expected an error")
			}
		})
	})
}

func withTmpDir(t *testing.T, input string, test func(dirname string)) {
	dirname, err := tmpDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirname)

	if err := writeDefaultSupfile(dirname, input); err != nil {
		t.Fatal(err)
	}

	test(dirname)
}

func withTmpDirWithoutSupfile(t *testing.T, test func(dirname string)) {
	dirname, err := tmpDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirname)

	test(dirname)
}

func tmpDir() (string, error) {
	return ioutil.TempDir("", "suptest")
}

func writeDefaultSupfile(dirname, input string) error {
	return writeSupfileAs(dirname, "Supfile", input)
}

func writeSupfileAs(dirname, filename, input string) error {
	return ioutil.WriteFile(
		filepath.Join(dirname, filename),
		[]byte(input),
		0666,
	)
}

func testEnv() venv.Env {
	env := venv.Mock()
	env.Setenv("USER", envTestUser)
	return env
}
