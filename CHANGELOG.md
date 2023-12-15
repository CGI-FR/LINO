# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Types of changes

- `Added` for new features.
- `Changed` for changes in existing functionality.
- `Deprecated` for soon-to-be removed features.
- `Removed` for now removed features.
- `Fixed` for any bug fixes.
- `Security` in case of vulnerabilities.

## [2.6.0]

- `Added` `sqlserver` support
- `Added` flags `sample-size`, `distinct`, `limit`, `where` and `table` to the `analyse` command, by default distinct values are not counted

## [2.5.0]

- `Added` command `analyse` to extract metrics from the database in YAML format

## [2.4.0]

- `Added` go-ora driver for oracle in replacement of old driver (remove technical prerequisite to install Oracle Instant Client)

## [2.3.0]

- `Added` flag `--pk-translation` allow update of primary keys, by giving a cache.jsonl file containing old and new values for a specific key
- `Added` if a JSON object contains a `__usingpk__` with a dictionary of key/values, it will be used by the push command to select the target record to update (push update only)
- `Added` where clauses (child and parent) in ingress descriptor to enable non-start table filtering
- `Fixed` oracle connector disable/enable contraints cascade

## [2.2.0]

- `Added` import property in `tables.yaml` allow to specify format of data to read from JSON stream (`string` (default), `numeric`, `binary` or `base64` (same), `datetime`, `timestamp`) therefore `import` and `export` now mirror each other (`import` is used at push and `export` at pull) but `import` still allow to specify the data type to pass to database driver (backward compatibility)
- `Added` import property in `tables.yaml` allow to specify a format AND a type at the same time with the `format(type)` syntax (e.g. `import: binary(int64)`)
- `Added` websocket connector with basic auth, supported schemes : ws, wss (BETA)
- `Added` flag `--stats` to generate a stat file or HTTP POST
- `Added` flag `--statsTemplate` to control the format of generated stats
- `Fixed` `lino push truncate` with `--table` option doesn't truncate table #123

## [2.1.0]

- `Added` export mode all in `tables.yaml` to export all columns even if some columns are defined in the columns property

## [2.0.1]

- `Fixed` Bad SQL update statement for oracle
- `Fixed` Reset statement after error during push #54
- `Fixed` Continue to close others row writers after error

## [2.0.0]

- `Changed` order of keys in output JSON lines will be alphabetical when pulling (without configuration in tables.yaml)
- `Added` configuration of export format / import type for columns in tables.yaml, see issue #33 for more information
- `Added` MariaDB/MySQL support (thanks to @joaking85)
- `Added` auto-select columns required by a relation but not exported in tables.yaml
- `Added` new commands to configure tables : add-column and remove-column
- `Added` New command to count lines in tables
- `Fixed` limit keyword on DB2 dialect
- `Fixed` Push truncate respect child/parent constraint order
- `Fixed` Push truncate will trigger only for attainable tables in the ingress descriptor - tables that are cut out will not be truncated
- `Fixed` push insert with mysql/mariadb connector now works properly with MySQL database
- `Fixed` push insert with mysql/mariadb connector will not update record if it exists

## [1.10.0]

- `Added` HTTP connector will now close/reopen request when commit size is reached

## [1.9.2]

- `Fixed` charset on Content-Type header when pushing to HTTP connector backend

## [1.9.1]

- `Fixed` some HTTP library doesn't support a body payload on a HTTP GET request, therefore the HTTP connector will pass the "filter" parameter through the headers in addition to the request body

## [1.9.0]

- `Added` new verb to extract, to get status and to update sequences

## [1.8.0]

- `Added` new parameter to pull only distinct values from the start table

## [1.7.0]

- `Added` new datasource type with string connection `http://...` LINO can pull/push data to an HTTP Endpoint API

## [1.6.0]

- `Added` option to change ingress-descriptor filename
- `Changed` update debian image to last stable (debian:stable-20210816-slim)

## [1.5.0]

- `Added` update Pimo to v1.8.0

## [1.4.0]

- `Added` statistics report for push and pull executions (thanks to @CapKicklee)
- `Changed` some info level logs to debug level in pull module

## [1.3.1]

- `Fixed` Revert convert JSON date to Oracle date format as a workaround for godror

## [1.3.0]

- `Added` flag to enable or disable coloring in output logs (--color [yes|no|auto])
- `Added` update Pimo to v1.6.1

## [1.2.1]

- `Fixed` Remove ENTRYPOINT and change CMD in oracle docker image

## [1.2.0]

- `Added` structured logs (debug & json format) (thanks to @CapKicklee)

## [1.1.2]

- `Fixed` extract composite primary keys for oracle
- `Fixed` protect columns names in insert statement

## [1.1.1]

- `Fixed` Missing where keyword for Oracle SQL Query

## [1.1.0]

- `Added` --where flag to use a raw sql where clause to filter rows of start table
- `Added` Oracle database support
- `Security` remove connection string from log

## [1.0.0]

- `Added` First public version released
