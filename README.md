<p align="center">
  <img src="docs/assets/konform-logo.png" width="180" alt="konform logo">
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/nzhussup/konform"><img src="https://pkg.go.dev/badge/github.com/nzhussup/konform.svg" alt="Go Reference"></a>
  <a href="https://github.com/nzhussup/konform/actions/workflows/ci.yml"><img src="https://github.com/nzhussup/konform/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://goreportcard.com/report/github.com/nzhussup/konform"><img src="https://goreportcard.com/badge/github.com/nzhussup/konform" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/nzhussup/konform" alt="License"></a>
</p>

# konform

`konform` is a schema-first configuration library for Go.
Define typed config structs once, then load from files and environment with defaults and validation.

## Why konform

Configuration in Go often ends up split across multiple libraries and custom glue code.
That usually leads to duplicated mapping logic, unclear precedence, and inconsistent validation.

`konform` keeps this explicit and predictable by using struct tags as the schema and a small loading API.

## Key features

- Schema-first configuration from typed Go structs
- Multiple sources: environment variables, YAML files, JSON files
- Defaults via struct tags
- Required-field validation
- Nested struct support
- Deterministic precedence through explicit source order
- Clear, human-friendly validation and decode errors

## Installation

```bash
go get github.com/nzhussup/konform
```

## Quick start

```go
package main

import (
	"fmt"
	"log"

	"github.com/nzhussup/konform"
)

type Config struct {
	Server struct {
		Host string `key:"server.host" default:"127.0.0.1"`
		Port int    `key:"server.port" default:"8080" env:"PORT"`
	}
	Database struct {
		URL string `key:"database.url" env:"DATABASE_URL" required:"true"`
	}
}

func main() {
	var cfg Config

	err := konform.Load(
		&cfg,
		konform.FromYAMLFile("config.yaml"),
		konform.FromEnv(),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", cfg)
}
```

## Precedence

`konform` applies values in this order:

```text
defaults < file < env
```

This behavior is controlled by the order of options passed to `Load`.
If multiple sources set the same field, the later source wins.

## Tags

- `key`: path used for YAML/JSON lookup (defaults to struct field path when omitted)
- `env`: environment variable name
- `default`: default value used when the field is zero-valued before source loading
- `required:"true"`: marks a field as required after all sources are applied

Note: this release uses `key` for file mapping. `conf` is not currently a supported tag.

## Examples

See runnable examples in [`examples/`](examples/):

- `examples/basic`
- `examples/env`
- `examples/yaml`
- `examples/json`

## Philosophy

- Minimal magic
- Explicit behavior
- Idiomatic Go APIs and errors

## Status

> pre-v1, API may evolve

## License

MIT. See [LICENSE](LICENSE).
