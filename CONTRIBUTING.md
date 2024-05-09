# Contributing

## Discuss significant changes

Before you invest a significant amount of time on a change, please create a discussion or issue describing your proposal. This will help to ensure your proposed change has a reasonable chance of being merged.

## Avoid dependencies

Adding a dependency is a big deal. While on occasion a new dependency may be accepted, the default answer to any change that adds a dependency is no.

## Set up development environment

Lys tests are run against an actual PostgreSQL database which you can create automatically on a Linux system as follows:
* find the file "internal/sql/ddl/create_users.sql" and run the contents as superuser on your local cluster
* copy lys_config_sample.toml to lys_config.toml
* enter your db superuser password in lys_config.toml
* ensure you don't currently have a database with the database name shown
* ensure /home/\<user\>/go/bin is in PATH (check with `echo $PATH`)
* enter `make testdb`. This will (re-)create the database and populate it with tables and data

## Run tests

* ensure CGO is enabled (check with `go env`: if CGO_ENABLED=0, then `sudo apt install build-essentials` then `go env -w CGO_ENABLED=1`)
* from the root directory, `make tests`