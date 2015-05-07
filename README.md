Stack Up
========

Stack Up is a simple deployment tool that performs given set of commands on multiple hosts in parallel.

    $ sup <network> <target/command>
    
    $ # eg.
    $ sup prod deploy

# Installation

    $ go get github.com/pressly/stackup/cmd/sup

# Supfile

- **env** - Environment variable.
- **networks** - Network is a group of hosts, eg. `dev`, `stg` and `prod`.
- **commands** - Command represents named set of commands to be run remotelly.
- **targets** - Target is an alias for one or more commands.

See the [example Supfile](./example/Supfile) to deploy example golang server to a multiple hosts (local/dev/stg/prod networks).

# License
Stack Up is licensed under the [MIT License](./LICENSE).

