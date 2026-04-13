# Changelog

All notable changes to this project will be documented in this file.

## v0.4.0 - 2026-04-13



### Documentation

- Update changelog for v0.3.1 [skip ci] [skip codeql]

- Add godoc for each public api

- Improve documentation and add version file


### Features

- Add new validators

- Add a possibility for byte loads

- Add .env file support

- Add env prefix support

- Add custom validators


### Refactoring

- Rename error,warn,off modes to inlclude Mode prefix


## v0.3.1 - 2026-04-08



### Documentation

- Update changelog for v0.3.0 [skip ci]

- Add documentation


### Features

- Add strict mode

- Add reporting feature


### Fixes

- Skip codeql on release workflow [skip ci] [skip codeql]


## v0.3.0 - 2026-04-08



### Documentation

- Update changelog for v0.2.0 [skip ci]


### Features

- Add supported validator logic

- Evolve and extend validation options

- Add linter

- Add unknown key suggestions and configurable mode for suggestions


### Fixes

- Fix ci to use v2 linter

- Fmt


## v0.2.0 - 2026-04-08



### Documentation

- Add code of conduct

- Update changelog for v0.1.3 [skip ci]


### Features

- Imrpove error logging and avoid fail fast on type cast errors


### Refactoring

- Restructure validate dir to allow clean addition of additional validators

- Rename validations to validators

- Refactor required tag to be validate, so that validate can include multiple validations


## v0.1.3 - 2026-04-07



### Fixes

- Update release yml


## v0.1.2 - 2026-04-07



### Documentation

- Update changelog for v0.1.1 [skip ci]

- Update codeowners and release documentation

- Add templates


### Fixes

- Fix gocyclo issues


## v0.1.1 - 2026-04-07



### Features

- Add release ci


## v0.1.0 - 2026-04-07



### Features

- Add MVP supporting env loads for config

- Refactor internal struct and add unit test coverage

- Add JSON support and improve decoding by adding special cases

- Improve nested error messages

- Add full suite unit tests

- Add readme, makefile and prepare everything to a pre-release


### Refactoring

- Move validate into internal folder

- Rename conf to key

- Rename package name to konform


