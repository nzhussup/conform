# Reference

## Load

```go
func Load(target any, opts ...Option) (*LoadReport, error)
```

- `target` must be a non-nil pointer to a struct
- returns `(*LoadReport, error)`

## Options

- `FromJSONFile(path string)`
- `FromYAMLFile(path string)`
- `FromTOMLFile(path string)`
- `FromJSONBytes(data []byte)`
- `FromYAMLBytes(data []byte)`
- `FromTOMLBytes(data []byte)`
- `FromEnv()`
- `WithUnknownKeySuggestionMode(mode UnknownKeySuggestionMode)`
- `Strict()`

## UnknownKeySuggestionMode

- `konform.Warn`
- `konform.Error`
- `konform.Off`

## Errors

Top-level exported errors:

- `konform.ErrInvalidTarget`
- `konform.ErrInvalidSchema`
- `konform.ErrDecode`
- `konform.ErrValidation`

Validation failures return `*konform.ValidationError`.

## Report types

- `type LoadReport struct { Entries []ReportEntry }`
- `func (r *LoadReport) Print(w io.Writer)`

- `type ReportEntry struct {
    Path   string
    Value  string
    Source string
  }`
