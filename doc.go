/*
 * Copyright (c) 2018, Jeremy Bingham (<jeremy@goiardi.gl>)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
spqr is a consul based user management utility that watches consul for updates on a key prefix for users to add, update, or disable on a system.

Usage

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

License

spqr is licensed under the terms of the Apache 2.0 License.

Placeholder for more extensive documentation.
*/
package main
