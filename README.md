Stack Up
========

Stack Up is a simple deployment tool that performs given set of commands on multiple hosts in parallel. It reads Supfile, a YAML configuration file, which defines networks (groups of hosts), commands and targets.

# Demo

Demo using the following [Supfile](./example/Supfile):

[![Sup](https://github.com/pressly/sup/blob/gif/asciinema.gif?raw=true)](https://asciinema.org/a/19742?autoplay=1)

# Installation

    $ go get -u github.com/pressly/sup/cmd/sup

# Usage

    $ sup [-f Supfile] [--only <regexp>] <network> <target/command> [...]

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

# Development

    fork it..

    $ make deps
    $ make build

# License

Licensed under the [MIT License](./LICENSE).

# Supfile Tips and Tricks

## Baked-In Variables

`sup` provides a few baked-in variables that make developing and
re-using Supfiles easier.

 - `SUP_NETWORK` the name of the network that the command was
   originally issued against:

 - `SUP_NONCE` the date and time of the original command line
   invocation. Useful for communicating a nonce across hosts in the
   network. Can be overridden with by setting the environment variable
   `SUP_NONCE`.

```yaml
commands:
  preparerelase:
    desc: Prepare release dir
    run: mkdir -p /app/rels/$SUP_NONCE/

  config:
    desc: Upload/test config file.
    upload:
      - src: ./example.$SUP_NETWORK.cfg
        dst: /app/rels/$SUP_NONCE/
    run: test -f /app/rels/$SUP_NONCE/example.$SUP_NETWORK.cfg
...
```
