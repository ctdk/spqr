spqr
====

`spqr` is a [consul](https://www.consul.io/)-based user management utility that watches consul for updates on a key prefix for users to add, update, or disable on a system.

OVERVIEW
--------

User and group information is stored in consul's key/value store in JSON documents, and a consul watch with spqr runs on the nodes to be managed. When a group that spqr is watching gets updated, the consul watch is notified and spqr is run with the updated group definition. It then fetches the user information out of consul and creates, updates, or disables the users as needed.

### Users

The user definition JSON is structured like so:

```
{
  "username": "baz",
  "full_name": "Baz McBazzick",
  "groups": ["sysadmins", "bazzers"],
  "primary_group": "users",
  "action": "create",
  "shell": "/bin/sh",
  "authorized_keys": [
    "ssh-rsa AAAAAAAAAAA baz@q.local"
  ]
}
```

The mandatory fields are `username` and `action`, although unless the user is being disabled `authorized_keys` is strongly recommended. Default values are filled in for `shell` (`/bin/bash`) and `full_name` (set to `username`), while the default value for `primary_group` depends on the OS defaults for user primary groups (generally, it's a group named after the user, but it may not always be the case). The `action` is either `"create"` or `"disable"`.

These user definitions need to be stored in consul with a key that matches `USER_KEY_PREFIX/<username>`. By default the user key prefix is `org/default/users`, so the example above would be stored in `org/default/users/baz`.

### Groups

**NB:** The groups discussed here are not the same as groups in the operating system. Having a user in a group called `ops` in `spqr` will not put the user in an OS group called `ops` on the system, unless you added it to the groups in the user definition.

The group definition JSON is structured like so:

```
{
  "members":
  [
    {
      "username": "foo",
      "status": "enabled"
    },
    {
      "username": "bar",
      "status": "enabled"
    },
    {
      "username": "baz",
      "status": "enabled"
    },
    {
      "username": "bill",
      "status": "disabled"
    }
  ]
}
```

There's rather less to the groups than there is to the user definitions. There's just an array of hashes named `members`, where each hash has `username` and `status` keys. The available statuses are `enabled` and `disabled`. 

While the only hard constraint with the key in consul for groups is that the group key must match the prefix (or name) the consul watch is watching on, a good convention to use is to use a key similar to the ones used with users along the lines of `org/default/groups/<group name>`.

### Disabling users

If a user has the action `create`, but their status in the group definition is `disabled`, or if they're enabled in the group but marked as `disable` in the user definition, the user will be disabled. A user that is marked to be disabled that does not already exist on the system will not be created.

When a user is disabled, their ssh authorized keys are removed, they are removed from all their secondary groups, their login shell is changed to `/sbin/nologin`, and their processes are all killed. Their home directories are not removed, and they're still members of their primary group.

USAGE
-----

spqr has several command line options when it's run:

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

On the command line, spqr needs to be run in the consul watch like this:

```
consul watch -type=keyprefix -prefix=<path/to/group> spqr [OPTIONS]
```

PLATFORMS
---------

Currently spqr only supports Linux. Other Unixes should be able to be made work without too much effort, but there hasn't been any work done on that front. There are stubs for Darwin, and a few for Windows, but they aren't functional yet.

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
