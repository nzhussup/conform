# Contributing to Konform

Thanks for your interest in contributing to Konform.

Konform aims to be a small, predictable, and production-ready Go configuration library. Contributions are welcome, but changes are reviewed carefully to keep the API clean, idiomatic, and stable.

## Before you contribute

Please open an issue first for:
- new features
- API changes
- behavioral changes
- large refactors

Small bug fixes, tests improvements, docs fixes, and examples can usually be submitted directly as a pull request.

## What Konform values

When contributing, keep these project principles in mind:

- idiomatic Go
- small and intentional API surface
- explicit behavior over magic
- predictable configuration loading and precedence
- strong error messages
- maintainability over clever abstractions
- backward compatibility where reasonable

## Development setup

Requirements:
- Go 1.24 or newer
- Make

Clone the repository and run:

```bash
make test
````

If linting is available in the project, also run:

```bash
make lint
make test
```

Please make sure all tests pass before opening a pull request.

## Project structure

The repository is organized roughly as follows:

* `konform.go` – public entry points
* `options.go` – public options API
* `error.go` – public errors
* `internal/schema` – schema discovery and metadata
* `internal/defaults` – default value application
* `internal/decode` – decoding raw source values into typed fields
* `internal/validate` – validation logic
* `internal/source/...` – configuration sources such as env, YAML, and JSON
* `examples/` – runnable usage examples

## Pull request guidelines

Please keep pull requests focused and small.

A good pull request should:

* solve one problem
* include or update tests
* include documentation updates when behavior changes
* avoid unrelated refactors
* preserve backward compatibility unless the change is explicitly discussed first

## Coding guidelines

Please follow these rules when contributing:

* prefer simple and readable code
* avoid unnecessary dependencies
* keep public APIs minimal
* keep internal abstractions justified and small
* write clear error messages
* add tests for new behavior and edge cases
* favor table-driven tests where appropriate
* avoid speculative features

## Tests

Every behavior change or bug fix should include tests where practical.

Please add tests close to the affected package, for example:

* schema changes → `internal/schema`
* decoding changes → `internal/decode`
* defaults behavior → `internal/defaults`
* source behavior → corresponding `internal/source/...` package

## Commit style

Conventional-style commits are appreciated, for example:

* `feat: add nested env prefix support`
* `fix: handle invalid duration default`
* `docs: clarify source precedence`
* `test: add coverage for required nested fields`

## Documentation

If your change affects user-facing behavior, please update one or more of:

* `README.md`
* examples in `examples/`
* `CHANGELOG.md` if maintainers ask for it

## Release process

Releases and version tags are managed by the maintainer.

On release tags, CI updates `version.go` with the tag version and commits it back to `main` together with changelog updates.

Please do not create tags or prepare releases in pull requests unless explicitly requested.

## Review process

Konform accepts open contributions, but the review process is intentionally strict.

Pull requests may be declined if they:

* expand the API without strong justification
* introduce hidden behavior or too much magic
* add maintenance burden without clear value
* conflict with the project's design goals

## Security

If you discover a security issue, please do not open a public issue.
Instead, report it privately through the repository security contact process if available.

## Questions

If you are unsure whether a change fits the project, open an issue first and describe:

* the problem
* the proposed solution
* tradeoffs
* expected user-facing behavior

Thanks for helping improve Konform.