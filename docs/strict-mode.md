# Strict Mode

Use strict mode when you want fast feedback for config mistakes.

```go
report, err := konform.Load(
	&cfg,
	konform.Strict(),
	konform.FromJSONFile("config.json"),
	konform.FromEnv(),
)
```

## What strict mode enforces

- unknown keys in JSON/YAML/TOML are errors
- decode/type mismatches are errors
- invalid mapped env values are errors
- duplicate/conflicting schema mappings are errors
  - same `key` used by multiple fields
  - same `env` used by multiple fields

## What strict mode does not enforce

- missing optional fields are not errors (zero/default stays)
- unrelated environment variables are ignored

To enforce presence, use validation tags like:

```go
Field string `validate:"required"`
```
