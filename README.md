<p align="center">
  <img src="docs/assets/konform-logo.png" width="300" alt="konform logo">
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/nzhussup/konform"><img src="https://pkg.go.dev/badge/github.com/nzhussup/konform.svg" alt="Go Reference"></a>
  <a href="https://github.com/nzhussup/konform/actions/workflows/ci.yml"><img src="https://github.com/nzhussup/konform/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://goreportcard.com/report/github.com/nzhussup/konform"><img src="https://goreportcard.com/badge/github.com/nzhussup/konform" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/nzhussup/konform" alt="License"></a>
</p>

# konform

`konform` is a schema-first configuration library for Go.
Define typed config structs once, then load from files and environment with defaults, validation, strict checks, and load reporting.

## Why konform

Configuration in Go often ends up split across multiple libraries and custom glue code.
That usually leads to duplicated mapping logic, unclear precedence, and inconsistent validation.

`konform` keeps this explicit and predictable by using struct tags as the schema and a small loading API.

## Konform vs Viper and Koanf

`konform` is not trying to be the most feature-rich configuration framework.
It is optimized for strict, typed, schema-first loading with explainable outcomes.

| Capability | Konform | Viper | Koanf |
|---|---|---|---|
| Schema-first from Go structs | Strong default | Possible, but often extra wiring | Possible, often parser/provider composition |
| Strict unknown-key handling with suggestions | Built-in (`Warn` / `Error` / `Off`) | Usually custom validation needed | Usually custom validation needed |
| Built-in load explainability report | Built-in (`LoadReport`) | Not a core built-in concept | Not a core built-in concept |
| Strict mapping conflict checks (`key`/`env`) | Built-in (`Strict`) | Usually manual | Usually manual |
| Secret masking in report | Built-in (`secret` tag) | Usually manual | Usually manual |
| Dynamic/reload-heavy provider ecosystem | Limited by design | Strong | Strong |

Use `konform` when you want predictable, typed config behavior and explicit failure modes.
Use Viper/Koanf when you need broader provider ecosystems, dynamic config workflows, or more runtime flexibility.

## Key features

- Schema-first configuration from typed Go structs
- Multiple sources: environment variables, YAML/JSON/TOML files, YAML/JSON/TOML bytes
- Defaults via struct tags
- Validation rules (`required`, `min`, `max`, `len`, `minlen`, `maxlen`, `regex`, `oneof`, `nonzero`, `url`, `email`)
- Strict mode (`konform.Strict()`) for unknown structured keys and mapping conflicts
- Unknown-key handling modes: `konform.Warn`, `konform.Error`, `konform.Off`
- Load report with field values and value sources
- Secret masking in reports using `secret:"true"`
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
	"os"

	"github.com/nzhussup/konform"
)

type Config struct {
	Server struct {
		Host string `key:"server.host" default:"127.0.0.1"`
		Port int    `key:"server.port" default:"8080" env:"PORT"`
	}
	Database struct {
		URL string `key:"database.url" env:"DATABASE_URL" validate:"required" secret:"true"`
	}
	Log struct {
		Level string `key:"log.level" default:"info"`
	}
}

func main() {
	var cfg Config

	report, err := konform.Load(
		&cfg,
		konform.FromYAMLFile("config.yaml"),
		konform.FromEnv(),
		konform.WithUnknownKeySuggestionMode(konform.Warn),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", cfg)
	report.Print(os.Stdout)
}
```

## Precedence

`konform` applies values in this order:

```text
defaults < file < env
```

This behavior is controlled by the order of options passed to `Load`.
If multiple sources set the same field, the later source wins.

## Strict mode

Use `konform.Strict()` to enforce:

- unknown keys in structured sources are errors
- decode/type mismatches are errors
- invalid mapped env values are errors
- duplicate/conflicting `key` or `env` mappings are errors
- unrelated environment variables are ignored

## Tags

- `key`: path used for YAML/JSON lookup (defaults to struct field path when omitted)
- `env`: environment variable name
- `default`: default value used when the field is zero-valued before source loading
- `validate`: validation rules after all sources are applied
- `secret`: if truthy (`true`, `1`, `yes`, `on`), report value is masked as `***`

## Unknown key modes

- `konform.Warn` (default): print warning, continue load
- `konform.Error`: return decode error
- `konform.Off`: ignore unknown keys

Set mode with:

```go
konform.WithUnknownKeySuggestionMode(konform.Error)
```

## Load report

`Load` returns `(*LoadReport, error)`.

Report entries include:
- resolved path
- final value (masked for `secret` fields)
- source (`default`, file path, `env:VAR_NAME`, or `zero`)

Print formatted report:

```go
report.Print(os.Stdout)
```

## Examples

See runnable examples in [`examples/`](examples/):

- `examples/basic`
- `examples/env`
- `examples/yaml`
- `examples/json`
- `examples/toml`
- `examples/report`

## Documentation

More docs:

- [Getting Started](docs/getting-started.md)
- [Features](docs/features.md)
- [Strict Mode](docs/strict-mode.md)
- [Reporting](docs/reporting.md)
- [Reference](docs/reference.md)

## Philosophy

- Minimal magic
- Explicit behavior
- Idiomatic Go APIs and errors

## Status

> pre-v1, API may evolve

## License

MIT. See [LICENSE](LICENSE).
