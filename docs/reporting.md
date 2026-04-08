# Reporting and Explainability

`Load` returns a report:

```go
report, err := konform.Load(&cfg, konform.FromYAMLFile("config.yaml"))
```

You can print it:

```go
report.Print(os.Stdout)
```

Example output:

```text
server.port   = 9090      source=config.yaml
server.host   = 0.0.0.0   source=default
database.url  = ***       source=env:DATABASE_URL
log.level     = debug     source=config.yaml
```

## Source values

Each entry has:

- `Path`: resolved lookup path
- `Value`: final field value (masked for secret fields)
- `Source`:
  - `default`
  - file path (for file sources)
  - `env:VAR_NAME`
  - `zero` (not set by default or sources)

## Secret masking

Mask a field in report with:

```go
Password string `env:"DB_PASSWORD" secret:"true"`
```

Truthy `secret` values: `true`, `1`, `yes`, `on`.
