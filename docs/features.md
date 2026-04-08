# Features

## Schema-first loading

Your Go struct is the schema. Tags define mapping and behavior.

## Supported sources

- JSON files (`FromJSONFile`)
- YAML files (`FromYAMLFile`)
- TOML files (`FromTOMLFile`)
- Environment variables (`FromEnv`)

## Tags

- `key`: structured source lookup path
- `env`: environment variable name
- `default`: default value applied before sources
- `validate`: validation rules after loading
- `secret`: mask field value in report (`***`)

## Validation

Validation runs after all defaults and sources are applied.

Common rules include:

- `required`
- `min`, `max`
- `len`, `minlen`, `maxlen`
- `regex`
- `oneof`

## Unknown key handling

Set with `WithUnknownKeySuggestionMode`:

- `konform.Warn` (default): warning, no error
- `konform.Error`: decode error
- `konform.Off`: ignore

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
