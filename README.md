Stack Up
========

Stack Up is a simple deployment tool that performs given set of commands on multiple hosts in parallel. It reads Supfile, a YAML configuration file, which defines networks (groups of hosts), commands and targets.

    $ sup <network> <target/command>
    
    $ # eg.
    $ sup prod deploy

[![Sup](./example/sup.png)](https://asciinema.org/a/19658)

# Installation

    $ go get github.com/pressly/stackup/cmd/sup

# Supfile

See the [example Supfile](./example/Supfile) to deploy example golang server to a multiple hosts (local/dev/stg/prod networks).

- **env** - Environment variable.
- **networks** - Network is a group of hosts, eg. `dev`, `stg` and `prod`.
- **commands** - Command represents named set of commands to be run remotelly.
- **targets** - Target is an alias for one or more commands.

# License
Stack Up is licensed under the [MIT License](./LICENSE).
