<!--
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# snapd Configuration File

snapd supports being configured through a configuration file located at a default location of `/etc/snap/snapd.conf` on Linux systems or by passing a configuration file in through the `--config` command line flag when starting snapd. YAML and JSON are currently supported for configuration file types.

snapd runs without a configuration file provided and will use the default values defined inside the daemon (shown below). There is an order of precedence when it come to default values, configuration files, and flags when snapd starts. Any value defined in the default configuration file located at `/etc/snap/snapd.conf` will take precedence over default values. Any value defined in a configuration file passed via the `--config` flag will be used in place of any default configuration file on the system and override default values. Any flags passed in on the command line during the start up of snapd will override any values defined in configuration files and default values.

In order of precedence (from greatest to least):
- Command-line flags
- Configuration file passed in via the `--config` flag
- Default configuration file (if exists)
- Default values per configuration setting

## Usage
The configuration file is comprised of different sections for each module that the Snap daemon can run. Settings specifically for the Snap daemon are defined on the top level, along with configuration sections for Control, Scheduler, REST API Server, and Tribe. Below, each section will be detailed in YAML format broken out for each section. A full example configuration file can be seen in YAML or JSON format in examples/configs in the project source.

## YAML Example
When defining a configuration in YAML format, options or sections can be commented out if the value provided will not be different from the default value configured by the system.

### snapd configuration
This section comprises of configuration settings that are specific for the Snap daemon.

```yaml
---
# log_level for the snap daemon. Supported values are
# 1 - Debug, 2 - Info, 3 - Warning, 4 - Error, 5 - Fatal.
# Default value is 3.
log_level: 3

# log_path sets the path for logs for the snap daemon. By
# default snapd prints all logs to stdout. Any provided
# path will send snapd logs to a file called snapd.log in
# the provided directory.
log_path: /var/log/snap

# log_truncate specifies how the log file with be opened
# false => append
# true  => truncate
log_truncate: false

# log_colors specifies if log file output is colorified
# true  => colors
# false => no colors
log_colors: true

# Gomaxprocs sets the number of cores to use on the system
# for snapd to use. Default for gomaxprocs is 1
gomaxprocs: 1
```

### snapd control configurations
The control section contains settings for configuring the Control module within the Snap daemon. These configuration settings are specific for the running of plugins and the plugins section under control allows for passing of plugin specific configuration items to the plugins during a task run.

```yaml
control:
  # auto_discover_path sets a directory to auto load plugins on the start
  # of the snap daemon
  auto_discover_path: /opt/snap/plugins

  # cache_expiration sets the time interval for the plugin cache to use before
  # expiring collection results from collect plugins. Default value is 500ms
  cache_expiration: 500ms

  # max_running_plugins sets the size of the available plugin pool for each
  # plugin loaded in the system. Default value is 3
  max_running_plugins: 3

  # plugin_load_timeout sets the maximal time allowed for a plugin to load
  # Default value is 3
  plugin_load_timeout: 10

  # keyring_paths sets the directory(s) to search for keyring files for signed
  # plugins. This can be a comma separated list of directories
  keyring_paths: /opt/snap/plugins/keyrings

  # plugin_trust_level sets the plugin trust level for snapd. The default state
  # for plugin trust level is enabled (1). When enabled, only signed plugins that can
  # be verified will be loaded into snapd. Signatures are verified from
  # keyring files specified in keyring_path. Plugin trust can be disabled (0) which
  # will allow loading of all plugins whether signed or not. The warning state allows
  # for loading of signed and unsigned plugins. Warning messages will be displayed if
  # an unsigned plugin is loaded. Any signed plugins that can not be verified will
  # not be loaded. Valid values are 0 - Off, 1 - Enabled, 2 - Warning
  plugin_trust_level: 1

  # plugins section contains plugin config settings that will be applied for
  # plugins across tasks.
  plugins:
    all:
      password: p@ssw0rd
    collector:
      all:
        user: jane
      pcm:
        all:
          path: /usr/local/pcm/bin
        versions:
          1:
            user: john
            someint: 1234
            somefloat: 3.14
            somebool: true
      psutil:
        all:
          path: /usr/local/bin/psutil
    publisher:
      influxdb:
        all:
          server: xyz.local
          password: $password
    processor:
      movingaverage:
        all:
          user: jane
        versions:
          1:
            user: tiffany
            password: new password
```

### snapd scheduler configurations
The scheduler section of the configuration file configures settings for the Scheduler module inside the Snap daemon.

```yaml
scheduler:
  # work_manager_queue_size sets the size of the worker queue inside snapd scheduler.
  # Default value is 25.
  work_manager_queue_size: 25

  # work_manager_pool_size sets the size of the worker pool inside snapd scheduler.
  # Default value is 4.
  work_manager_pool_size: 4
```

### snapd REST API configurations
The restapi section of the configuration file configures settings for enabling and running the REST API as part of the Snap daemon. The snapctl command line tool uses the REST API to manage snapd on a host.

```yaml
restapi:
  # enable controls enabling or disabling the REST API for snapd. Default value is enabled.
  enable: true

  # https enables HTTPS for the REST API. If no default certificate and key are provided, then
  # the REST API will generate a private and public key to use for communication. Default
  # value is false
  https: false

  # rest_auth enables authentication for the REST API. Default value is false
  rest_auth: false

  # rest_auth_password sets the password to use the REST API. Currently user and password
  # combinations are not supported.
  rest_auth_password: changeme

  # rest_certificate is the path to the certificate to use for REST API when HTTPS is also enabled.
  rest_certificate: /etc/snap/certs/snap.pub

  # rest_key is the path to the private key for the certificate in use by the REST API
  # when HTTPs is enabled.
  rest_key: /etc/snap/certs/snap.key

  # port sets the port to start the REST API server on. Default is 8181
  port: 8181
```

### snapd tribe configurations
The tribe section of the configuration file configures settings for enabling and running tribe as part of the Snap daemon.
```yaml
tribe:
  # enable controls enabling tribe for the snapd instance. Default value is false.
  enable: false

  # bind_addr sets the IP address for tribe to bind.
  bind_addr: 0.0.0.0

  # bind_port sets the port for tribe to listen on. Default value is 6000
  bind_port: 6000

  # name sets the name to use for this snapd instance in the tribe
  # membership. Default value defaults to the local hostname of the system.
  name: snaphost-01

  # seed sets the snapd instance to use as the seed for tribe communications
  seed: 192.168.1.2:6000
```

## JSON Example
The same configuration settings above can also be provided in a JSON formatted configuration file. Unlike YAML which allows for commenting out unused options or whole sections, those unused options and/or sections are just removed from the JSON file.

```json
{
    "log_level": 1,
    "control": {
        "cache_expiration": "1s",
        "plugin_trust_level": 0
    },
    "restapi": {
        "enable": true,
        "https": true,
        "port": 8282
    }
}
```

## Restarting snapd to pick up configuration changes
If changes are made to the configuration file, `snapd` must be restarted to pick up those changes. Fortunately, this is a simple matter of sending a `SIGHUP` signal to the `snapd` process. For example, the following command will restart the `snapd` process on the local system:

```bash
$ kill -HUP `pidof snapd`
```

Note that in this example, we are using the `pidof` command to retrieve the process ID of the `snapd` process. If the `pidof` command is not available on your system you might have to use a `ps aux` command and pipe the output of that command to a `grep snapd` command in order to obtain the process ID of the `snapd` process. Once the `snapd` process receives that signal it will restart and pick up any changes that have been made to the configuration file that was originally used to Âstart the `snapd` process.

Do keep in mind that this signal will trigger a **restart** of the `snapd` process. This means that any running tasks will be shut down and any loaded plugins will be unloaded. In reality, this means that when the `snapd` process restarts any plugins not in the `auto_discover_path` will need to be loaded manually once the `snapd` process restarts (and any tasks not in that same `auto_discover_path` will need to be restarted). However, any plugins in the `auto_discover_path` will be automatically reloaded and any tasks in that same `auto_discover_path` will be automatically restarted when the when the `snapd` process restarts in response to a `SIGHUP` signal.

## More information
* [SNAPD.md](SNAPD.md)
* [REST_API.md](REST_API.md)
* [PLUGIN_SIGNING.md](PLUGIN_SIGNING.md)
* [TRIBE.md](TRIBE.md)
