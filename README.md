Stack Up
========

# or just stackup/cmd/sup   ...?

or `sup`/`stack`/`st` for short.

# Stackfile

...... what language bindings to execute stuff...?
.. mimic the shell.. whatever shell can do, this thing should be able to do..

# Supfile

See [example Supfile](./Supfile).

# Usage

    $ sup <host-group> <command-alias>

    $ sup prod deploy
    $ sup prod deploy --service=hubserver

    $ sup stg stop --service=hubserver
    $ sup stg stop -s=hubserver

    $ sup service build
    $ sup service deploy -s=hubserver # .. does it all......

    $ sup host top
    $ sup health
    $ sup stats

    $ sup exec ls -la

