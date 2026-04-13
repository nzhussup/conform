# Features

## Schema-first loading

Your Go struct is the schema. Tags define mapping and behavior.

## Supported sources

- JSON files (`FromJSONFile`)
- YAML files (`FromYAMLFile`)
- TOML files (`FromTOMLFile`)
- .env files (`FromDotEnvFile`)
- JSON bytes (`FromJSONBytes`)
- YAML bytes (`FromYAMLBytes`)
- TOML bytes (`FromTOMLBytes`)
- Environment variables (`FromEnv`)

### Environment prefixing

Use `WithEnvPrefix` to prepend a prefix for all `env` tag lookups.

- Applies to both `FromEnv` and `FromDotEnvFile`
- Useful when your application namespaces variables (for example: `APP_`)

```go
konform.Load(
	&cfg,
	konform.FromDotEnvFile(".env"),
	konform.FromEnv(),
	konform.WithEnvPrefix("APP_"),
)
```

## Tags

- `key`: structured source lookup path
- `env`: environment variable name
- `default`: default value applied before sources
- `validate`: validation rules after loading
- `secret`: mask field value in report (`***`)

## Custom type decoding

Fields that implement `encoding.TextUnmarshaler` are decoded from strings.

- Works for both value and pointer fields
- Applies consistently across file, `.env`, and environment-variable sources

```go
type LogFormat string

func (f *LogFormat) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "json", "text":
		*f = LogFormat(strings.ToLower(string(text)))
		return nil
	default:
		return fmt.Errorf("invalid format %q", string(text))
	}
}

type Config struct {
	Format *LogFormat `key:"log.format" env:"LOG_FORMAT"`
}
```

## Validation

Validation runs after all defaults and sources are applied.

Common rules include:

- `required`
- `min`, `max`
- `len`, `minlen`, `maxlen`
- `regex`
- `oneof`
- `nonzero`
- `url`
- `email`

### Custom validators

Register domain-specific validation rules with `WithCustomValidator` and use them in `validate` tags.

- Signature: `func(value any, ruleValue string) error`
- `ruleValue` comes from the tag argument (for example: `validate:"startswith=svc-"`)
- Return `nil` when valid, or an `error` to fail validation

```go
startsWith := func(value any, ruleValue string) error {
	raw, ok := value.(string)
	if !ok {
		return errors.New("expected string")
	}
	if !strings.HasPrefix(raw, ruleValue) {
		return fmt.Errorf("must start with %q", ruleValue)
	}
	return nil
}

konform.Load(
	&cfg,
	konform.WithCustomValidator("startswith", startsWith),
)
```

## Unknown key handling

Set with `WithUnknownKeySuggestionMode`:

- `konform.ModeWarn` (default): warning, no error
- `konform.ModeError`: decode error
- `konform.ModeOff`: ignore

Suggestions are schema-first:

- unexpected config key -> closest schema key

## Strict mode

`konform.Strict()` enables stricter guarantees:

- unknown keys in structured sources are errors
- decode/type mismatches are errors
- invalid mapped env values are errors
- duplicate/conflicting `key` mappings are errors
- duplicate/conflicting `env` mappings are errors
- unrelated env vars are ignored

## Why choose Konform over Viper/Koanf

Choose `konform` when you want:

- schema-first config from Go structs as the default model
- strict unknown-key behavior with built-in suggestions
- built-in explainability report (`LoadReport`) for auditing final values and sources
- built-in strict conflict checks for mapping tags

Choose Viper/Koanf when you need broader runtime/provider flexibility and dynamic configuration patterns over strict schema-centric loading.
