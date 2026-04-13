package konform

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nzhussup/konform/internal/errs"
	internalschema "github.com/nzhussup/konform/internal/schema"
	envsource "github.com/nzhussup/konform/internal/source/env"
	jsonsource "github.com/nzhussup/konform/internal/source/json"
	tomlsource "github.com/nzhussup/konform/internal/source/toml"
	yamlsource "github.com/nzhussup/konform/internal/source/yaml"

	"github.com/nzhussup/konform/internal/source/common"
)

// Option configures Load behavior.
type Option func(*loadOptions) error

type sourceLoader func(*internalschema.Schema) error

type loadOptions struct {
	sources               []sourceLoader
	unknownKeySuggestMode common.UnknownKeySuggestionMode
	envPrefix             string
	customValidators      map[string]CustomValidatorFunc
	strict                bool
}

// UnknownKeySuggestionMode controls how unknown structured keys are handled.
type UnknownKeySuggestionMode = common.UnknownKeySuggestionMode

const (
	// Warn prints a warning for unknown keys and continues.
	Warn UnknownKeySuggestionMode = common.UnknownKeySuggestionWarn
	// Error returns a decode error for unknown keys.
	Error UnknownKeySuggestionMode = common.UnknownKeySuggestionError
	// Off ignores unknown keys.
	Off UnknownKeySuggestionMode = common.UnknownKeySuggestionOff
)

type fileSourceFactory func(path string, callerDir string, suggestionMode common.UnknownKeySuggestionMode, o *loadOptions) sourceLoader
type bytesSourceFactory func(data []byte, suggestionMode common.UnknownKeySuggestionMode) sourceLoader

// CustomValidatorFunc validates a field value using the optional rule value
// from the validate tag (for example: validate:"myrule=arg").
//
// Return nil when valid, or a non-nil error to report a validation failure.
type CustomValidatorFunc func(value any, ruleValue string) error

// FromEnv loads values from process environment variables using field env tags.
func FromEnv() Option {
	return func(o *loadOptions) error {
		if o == nil {
			return errs.InvalidSchemaNilOptions
		}

		o.sources = append(o.sources, func(sc *internalschema.Schema) error {
			return envsource.LoadWithPrefix(sc, o.envPrefix)
		})
		return nil
	}
}

// FromDotEnvFile loads values from a .env file using field env tags.
func FromDotEnvFile(path string) Option {
	return fromFile(path, errs.InvalidSchemaEmptyDotEnv, func(path string, callerDir string, _ common.UnknownKeySuggestionMode, o *loadOptions) sourceLoader {
		source := envsource.NewDotEnvFileSource(path, callerDir, o.envPrefix)
		return source.LoadFile
	})
}

// FromYAMLFile loads values from a YAML file.
func FromYAMLFile(path string) Option {
	return fromFile(path, errs.InvalidSchemaEmptyYAML, func(path string, callerDir string, suggestionMode common.UnknownKeySuggestionMode, _ *loadOptions) sourceLoader {
		source := yamlsource.NewFileSource(path, callerDir, suggestionMode)
		return source.LoadFile
	})
}

// FromJSONFile loads values from a JSON file.
func FromJSONFile(path string) Option {
	return fromFile(path, errs.InvalidSchemaEmptyJSON, func(path string, callerDir string, suggestionMode common.UnknownKeySuggestionMode, _ *loadOptions) sourceLoader {
		source := jsonsource.NewFileSource(path, callerDir, suggestionMode)
		return source.LoadFile
	})
}

// FromTOMLFile loads values from a TOML file.
func FromTOMLFile(path string) Option {
	return fromFile(path, errs.InvalidSchemaEmptyTOML, func(path string, callerDir string, suggestionMode common.UnknownKeySuggestionMode, _ *loadOptions) sourceLoader {
		source := tomlsource.NewFileSource(path, callerDir, suggestionMode)
		return source.LoadFile
	})
}

// FromYAMLBytes loads values from YAML bytes.
func FromYAMLBytes(data []byte) Option {
	return fromBytes(data, errs.InvalidSchemaEmptyYAMLBytes, func(data []byte, suggestionMode common.UnknownKeySuggestionMode) sourceLoader {
		source := yamlsource.NewByteSource(data, suggestionMode)
		return source.LoadBytes
	})
}

// FromJSONBytes loads values from JSON bytes.
func FromJSONBytes(data []byte) Option {
	return fromBytes(data, errs.InvalidSchemaEmptyJSONBytes, func(data []byte, suggestionMode common.UnknownKeySuggestionMode) sourceLoader {
		source := jsonsource.NewByteSource(data, suggestionMode)
		return source.LoadBytes
	})
}

// FromTOMLBytes loads values from TOML bytes.
func FromTOMLBytes(data []byte) Option {
	return fromBytes(data, errs.InvalidSchemaEmptyTOMLBytes, func(data []byte, suggestionMode common.UnknownKeySuggestionMode) sourceLoader {
		source := tomlsource.NewByteSource(data, suggestionMode)
		return source.LoadBytes
	})
}

func fromFile(path string, emptyPathErr error, factory fileSourceFactory) Option {
	if path == "" {
		return func(o *loadOptions) error {
			return emptyPathErr
		}
	}

	callerDir := callerDirectory(3)
	return func(o *loadOptions) error {
		if o == nil {
			return errs.InvalidSchemaNilOptions
		}

		o.sources = append(o.sources, func(sc *internalschema.Schema) error {
			mode := o.unknownKeySuggestMode
			if o.strict {
				mode = common.UnknownKeySuggestionError
			}
			load := factory(path, callerDir, mode, o)
			return load(sc)
		})
		return nil
	}
}

func fromBytes(data []byte, emptyDataErr error, factory bytesSourceFactory) Option {
	if len(data) == 0 {
		return func(o *loadOptions) error {
			return emptyDataErr
		}
	}

	// Copy once to keep option input immutable for callers.
	dataCopy := append([]byte(nil), data...)

	return func(o *loadOptions) error {
		if o == nil {
			return errs.InvalidSchemaNilOptions
		}

		o.sources = append(o.sources, func(sc *internalschema.Schema) error {
			mode := o.unknownKeySuggestMode
			if o.strict {
				mode = common.UnknownKeySuggestionError
			}
			load := factory(dataCopy, mode)
			return load(sc)
		})
		return nil
	}
}

// WithUnknownKeySuggestionMode sets handling mode for unknown structured keys.
func WithUnknownKeySuggestionMode(mode UnknownKeySuggestionMode) Option {
	return func(o *loadOptions) error {
		if o == nil {
			return errs.InvalidSchemaNilOptions
		}

		o.unknownKeySuggestMode = mode
		return nil
	}
}

// Strict enables strict schema loading and mapping checks.
func Strict() Option {
	return func(o *loadOptions) error {
		if o == nil {
			return errs.InvalidSchemaNilOptions
		}

		o.strict = true
		return nil
	}
}

// WithEnvPrefix prepends prefix to all env-tag lookups for env and .env sources.
func WithEnvPrefix(prefix string) Option {
	return func(o *loadOptions) error {
		if o == nil {
			return errs.InvalidSchemaNilOptions
		}

		o.envPrefix = prefix
		return nil
	}
}

// WithCustomValidator registers a custom validator by rule name.
// The rule can then be used in validate tags (for example: validate:"myrule=arg").
func WithCustomValidator(name string, fn CustomValidatorFunc) Option {
	name = strings.TrimSpace(name)
	if name == "" {
		return func(o *loadOptions) error {
			return fmt.Errorf("%w: custom validator name must not be empty", errs.InvalidSchema)
		}
	}
	if fn == nil {
		return func(o *loadOptions) error {
			return fmt.Errorf("%w: custom validator %q must not be nil", errs.InvalidSchema, name)
		}
	}

	return func(o *loadOptions) error {
		if o == nil {
			return errs.InvalidSchemaNilOptions
		}

		if o.customValidators == nil {
			o.customValidators = make(map[string]CustomValidatorFunc)
		}
		o.customValidators[name] = fn
		return nil
	}
}

func callerDirectory(skip int) string {
	_, filename, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	return filepath.Dir(filename)
}
