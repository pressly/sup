Stack Up
========

# or just stackup/cmd/sup   ...?

or `sup` for short.

# Stackfile

...... what language bindings to execute stuff...?
.. mimic the shell.. whatever shell can do, this thing should be able to do..

# Usage

$ sup deploy
$ sup deploy --service=hubserver

$ sup stop --service=hubserver
$ sup stop -s=hubserver

$ sup service:deploy

$ sup app:deploy
$ sup app:stop
$ sup app:restart
$ sup app:build

$ sup service:build
$ sup service:deploy -s=hubserver # .. does it all......

$ sup top
$ sup health
$ sup stats

$ sup exec ls -la

