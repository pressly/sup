Stack Up
========

Stack Up is a simple deployment tool that performs given set of commands on multiple hosts in parallel. It reads Supfile, a YAML configuration file, which defines networks (groups of hosts), commands and targets.

# Demo

Demo using the following [Supfile](./example/Supfile):

[![Sup](https://github.com/pressly/sup/blob/gif/asciinema.gif?raw=true)](https://asciinema.org/a/19742?autoplay=1)

# Installation

    $ go get -u github.com/pressly/sup/cmd/sup

# Usage

    $ sup [OPTIONS] NETWORK TARGET/COMMAND [...]

### Options

| Option            | Description                                  |
|-------------------|----------------------------------------------|
| `--help`, `-h`    | Print help/usage                             |
| `--version`, `-v` | Print version                                |
| `-f Supfile`      | Location of Supfile                          |
| `--only REGEXP`   | Filter NETWORK hosts using regexp string     |
| `--except REGEXP` | Filter out NETWORK hosts using regexp string |

### Network

A group of hosts on which COMMAND will be invoked in parallel.

```yaml
# Supfile

networks:
    production:
        hosts:
            - api1.example.com
            - api2.example.com
            - api3.example.com
    staging:
        hosts:
            - stg1.example.com
```

`$ sup production COMMAND` will invoke COMMAND on all production hosts in parallel.

`$ sup staging TARGET` will invoke TARGET on the staging host.

### Target

An alias to run multiple COMMANDS.

```yaml
# Supfile

targets:
    deploy:
        - build
        - pull
        - migrate-db-up
        - stop-rm-run
        - health
        - slack-notify
        - airbrake-notify
```

`$ sup production deploy` will invoke `build`, `pull`, `migrate-db-up`, `stop-rm-run` and `slack-notify` commands sequentially on all production hosts.

### Command

A shell command (or set of commands) to be run remotely.

```yaml
# Supfile

commands:
    restart:
        desc: Restart example Docker container
        run: sudo docker restart example
    tail-logs:
        desc: Watch tail of Docker logs from all hosts
        run: sudo docker logs --tail=20 -f example
    exec:
        desc: Exec into Docker container on all hosts
        stdin: true
        run: sudo docker exec -i example bash
    bash:
        desc: Interactive Bash on all hosts
        stdin: true
        run: bash
```

`$ sup production restart` will restart all production `example` Docker containers in parallel.

`$ sup production tail-logs` will tail Docker logs from all production `example` containers in parallel.

`$ sup production exec` will Docker Exec into all production Docker containers and run interactive shell.

`$ sup production bash` will run interactive shell on all production hosts.

# Supfile

See [example Supfile](./example/Supfile).

### Basic structure

```yaml
# Supfile
---
version: 0.3

# Global environment variables
env:
  NAME: api
  IMAGE: example/api

networks:
  local:
    hosts:
      - localhost
  staging:
    hosts:
      - stg1.example.com
  production:
    hosts:
      - api1.example.com
      - api2.example.com

commands:
  echo:
    desc: Print some env vars
    run: echo $NAME $IMAGE $SUP_NETWORK
  date:
    desc: Print OS name and current date/time
    run: uname -a; date

targets:
  all:
    - echo
    - date
```

### Default environment variables

- `$SUP_NETWORK` - Name of the NETWORK that the command was originally issued against.
- `$SUP_USER` - Name of user who issued the command.
- `$SUP_TIME` - Date and time of the original command line invocation.

# Development

    fork it..

    $ make tools
    $ make build

# License

Licensed under the [MIT License](./LICENSE).
