# Qvain Setup Guide

## Installation

### Prerequisites

Qvain has an [Ops repository](https://github.com/CSCfi/qvain-ops) that should be able to build a working system by means of the [Ansible](http://docs.ansible.com/) automation agent.

Here is a very rough outline of the minimally required steps:

- make sure the system is configured with a unicode locale;
- install Postgresql and configure to listen on unix socket;
- install Redis and configure to listen on unix socket;
- add a database and user for Qvain;
- install the Go programming language;
- ... profit!

### Build from source

These are the commands that build the Go backend. In short, you clone the repository, run `go get` to download the dependencies, and run `make` to build all library and command packages.

```shell
git clone qvain-api.git
cd qvain-api
go get -u ./...
make all
```

The makefile is a thin wrapper around standard go commands to link version information into the generated binaries. Instead of using the makefile in the last line above, you can also build the software using the standard idiomatic `go build` or `go install` commands:

```shell
env GOBIN=./bin go install ./...
```

... but the binaries will not have version information.

> **Note:** The ellipsis syntax `./...` means _all Go packages in this and subdirectories_. ([more](https://golang.org/cmd/go/#hdr-Package_lists))

To update the code, just pull in any changes and rebuild:

```shell
git pull -v
make all
```

Build times should be pretty fast: a few seconds for versions of Go before 1.10 and less than a second for later versions which cache and do incremental builds by default.

### Partial build

Most of the top-level directories are packages, i.e. library modules. The `cmd/` directory contains the actual commands that pull in those and external library packages and get built into self-contained binaries.

You can build a single command by providing its package name – its directory – as argument to make:

```shell
make qvain-backend
```

... or the standard go way:

```shell
go build -o bin/qvain-backend cmd/qvain-backend/
```

(But see the note about versioning information above if you don't use the makefile.)

## Configuration

Qvain uses gets its configuration from the environment. It is a good idea to create an env file with the needed variables – say `~/.env/qvain.env`. This file can be sourced at the beginning of a development session with `source ~/.env/qvain.env`, added to the app user's `bashrc` or included in a systemd unit file.

### Environment variables

These are the environment variables Qvain looks for. Those variables starting with "`PG`" are used to configure Postgresql connections; check the official [Postgresql documentation](https://www.postgresql.org/docs/9.6/static/libpq-envars.html) for more information.

| variable        | type      | description |
| --------------- | --------  | ----------- |
| `APP_DEBUG`     | `boolean` | log debugging statements; enable for development, not useful for production systems |
| `APP_INSTALLED` | `boolean` | is the application installed system-wide or running from its source repository |
| `TESTVAR`       | `string`  | test variable to check if environment has been set |
|                 |           | |
| `PGHOST`        |           | psql host name |
| `PGDATABASE`    |           | psql database name |
| `PGUSER`        |           | psql user name |
| `PGPASS`        |           | psql user password |
| `PGSSLMODE`     |           | psql ssl connection setting |
| `PGAPPNAME`     |           | psql application name |

Boolean values can be `0`, `false`, `no` or unset for *false*; everything else is *true*.

### Defaults

For performance and security reasons, it is preferred to run Postgresql and Redis from local Unix sockets instead of over TCP.


## Run-time

### Running backend services

TODO!!!

The backend includes a *systemd* unit file that should be enabled to start the service automatically. It is advised to start, stop or restart the service on production systems by means of that systemd unit file.

The service file makes sure that...
- Qvain starts after Postgresql and Redis;
- output from `STDOUT` is redirected to something that can save the logs;
- the proper environment variables are set.

When running the backend manually, those are the requirements you should pay attention to.

If you are developing, you can simply run the built binaries as they will print output or logs to console.

### Running commands

You can run any included command line utilities directly. Those commands that use the database need to have the Postgresql environment variables set so they know how to connect. The preferred way is to simply source the env configuration file:

```shell
source ~/.env/qvain.env
bin/some-cli-command -flag arg1 arg2
```

### Logging

Backend services write logs to `STDOUT` in JSON format; it's up to the administrator to do something with that output, such as redirecting to a file or piping to a log collecting tool.

The backend really only distinguishes between debugging and normal logging output. When running in production, the output should be relatively minimal – probably a few lines per request. In debugging mode, the output includes any debugging statements the developer has added to facilitate development; it is encouraged to remove (most) debugging log calls before releasing a production version of the software.

Production systems should not enable debugging output by default as the contents of these statements is really only meant to be useful for developers.

### Errors and Crashes

If Qvain encounters a fatal error on startup, it will write an error to `STDERR` and exit. Reasons for such fatal errors would be missing templates, SSL certificates or other filesystem related existence or permission problems. These problems are most likely to occur during installation or major updates; if Qvain has run successfully before, all file dependencies should be in place.

Once the backend is up and running, it does whatever it can to keep servicing requests. In case of crashes (a *panic* in Go), there is a panic handler that should catch crashing request handlers. The end-user or API client will most likely see a `500 Internal Server Error` page for that request but the server will otherwise keep on serving requests. The most likely reason for run-time crashes is problems with database connections, especially in the form of database methods being called on nil connection cursors.
