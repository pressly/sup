Stack Up
========

Stack Up is a simple deployment tool that performs given set of commands on multiple hosts in parallel. It reads Supfile, a YAML configuration file, which defines networks (groups of hosts), commands and targets.

    $ sup [-f Supfile] [--only <regexp>] <network> <target/command>

| command/option    | description                                                                                          |
|-------------------|------------------------------------------------------------------------------------------------------|
| `<network>`       | Group of hosts, eg. `production` that consists of 1-N hosts.                                         |
| `<command>`       | Name set of bash commands to be run remotely, eg `build` that triggers `docker build -t my/image .`. |
| `<target>`        | An alias to run multiple `<commands>`, eg `deploy` that triggers `pull`, `build` and `run` commands. |
|                   |                                                                                                      |
| `-f Supfile`      | Custom deployment config file (YAML) for `sup`, see [example Supfile](./example/Supfile).            |
| `--only <regexp>` | Filter `<target>` hosts by regexp string, eg `--only host1`.                                         |

## Examples:

    $ sup prod deploy
    $ sup --only api1 dev tail-logs
    $ sup -f Supfile.db stg restart

# Installation

    $ go get -u github.com/pressly/sup/cmd/sup

# Demo

[![Sup](https://github.com/pressly/sup/blob/gif/asciinema.gif?raw=true)](https://asciinema.org/a/19742?autoplay=1)

# Development

    fork it..

    $ make deps
    $ make build

# License
Licensed under the [MIT License](./LICENSE).
