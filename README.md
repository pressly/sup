Stack Up
========

Stack Up is a simple deployment tool that performs given set of commands on multiple hosts in parallel. It reads Supfile, a YAML configuration file, which defines networks (groups of hosts), commands and targets.

    $ sup <network> <target/command>

- `<network>` - A group of hosts, eg. `dev`, `stg` and `prod`. In this case, `prod` can map to `prod1.example.com`, `prod2.example.com` and `prod3.example.com` hosts.
- `<command>` - A named command (or set of commands) to be run remotely.
- `<target>` - An alias to run multiple `<commands>`.

`sup` picks up `Supfile` config file (YAML) from the current directory (the same way as `make` picks up `Makefile`). See [example Supfile](./example/Supfile).

[![Sup](./example/sup.png)](https://asciinema.org/a/19658)

# Installation

    $ go get github.com/pressly/stackup/cmd/sup

# License
Stack Up is licensed under the [MIT License](./LICENSE).
