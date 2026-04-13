# Stability and Support Policy

## Versioning

Konform follows Semantic Versioning (`MAJOR.MINOR.PATCH`).

- `MAJOR`: breaking API or behavior changes
- `MINOR`: backward-compatible features
- `PATCH`: backward-compatible fixes and security updates

## API compatibility

For stable major versions (`v1+`):

- Exported API changes are backward-compatible within the same major version.
- Breaking API changes are introduced only in a new major release.
- Internal packages (`internal/...`) are not part of the public compatibility contract.

## Behavior compatibility

Within a stable major version:

- Existing documented behavior should remain stable.
- Bug fixes may tighten invalid input handling, but should not break valid existing usage.
- Error message wording can evolve for clarity, but error categories and wrapping contracts are preserved.

## Deprecation policy

When deprecating public API:

- API is marked deprecated in doc comments.
- A replacement is documented.
- Deprecated API remains available until the next major release unless there is a critical security reason.

## Support policy

Konform maintainers provide regular fixes for:

- the latest stable release
- the previous stable release line when fixes are low-risk to backport

Security fixes are prioritized for the latest stable release.

## Release policy

- Releases are cut from Git tags (`v*`).
- Changelog and `version.go` are updated by release automation.
- Every release should pass CI (format, lint, tests, race tests).
