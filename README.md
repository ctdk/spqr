spqr
====

`spqr` is a consul based user management utility that watches consul for updates on a key prefix for users to add, update, or disable on a system.

USAGE
-----

```
Usage:
  spqr [OPTIONS]

Application Options:
  -v, --version           Print version info.
  -c, --config=           Specify a config file to use. [$SPQR_CONFIG_FILE]
  -C, --consul-http-addr= Consul HTTP API endpoint. Defaults to
                          http://127.0.0.1:8500. Shares the same
                          CONSUL_HTTP_ADDR environment variable as consul
                          itself as well. [$CONSUL_HTTP_ADDR]
  -P, --user-key-prefix=  Consul key prefix for user data. Default value:
                          'org/default/users'. [$SPQR_USER_KEY_PREFIX]
  -L, --log-file=         Log to file X [$SPQR_LOG_FILE]
  -S, --syslog            Log to syslog rather than to a log file. Incompatible
                          with -L/--log-file. [$SPQR_SYSLOG]
  -g, --log-level=        Specify logging verbosity.  Performs the same
                          function as -V, but works like the 'log-level' option
                          in the configuration file. Acceptable values are
                          'debug', 'info', 'warning', 'error', 'critical', and
                          'fatal'. [$SPQR_LOG_LEVEL]
  -s, --statefile=        Store spqr's state in this file.
  -V, --verbose           Show verbose debug information. Repeat for more
                          verbosity.

Help Options:
  -h, --help              Show this help message
```

TODO
----

There's quite a bit. See the TODO file for a hopefully up-to-date list.

BUGS
----

Presumably.

AUTHOR
------

Jeremy Bingham (<jeremy@goiardi.gl>)

COPYRIGHT
---------

Copyright 2018, Jeremy Bingham.

LICENSE
-------

`spqr` is licensed under the terms of the Apache 2.0 License.

EXPLANATION OF NAME
-------------------

[SPQR](https://en.wikipedia.org/wiki/SPQR) stands for *Senatus Populusque Romanus*, or "The Senate and People of Rome", and was the name of the Roman state. Since the yearly elected dual heads of state of the Roman Republic were consuls, `spqr` seemed like an appropriate name for a utility leveraging consul for its work.
