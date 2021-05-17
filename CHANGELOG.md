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
