package konform

import (
	"path/filepath"
	"runtime"

	"github.com/nzhussup/konform/internal/errs"
	internalschema "github.com/nzhussup/konform/internal/schema"
	envsource "github.com/nzhussup/konform/internal/source/env"
	jsonsource "github.com/nzhussup/konform/internal/source/json"
	tomlsource "github.com/nzhussup/konform/internal/source/toml"
	yamlsource "github.com/nzhussup/konform/internal/source/yaml"

	"github.com/nzhussup/konform/internal/source/common"
)

type Option func(*loadOptions) error

type sourceLoader func(*internalschema.Schema) error

type loadOptions struct {
	sources               []sourceLoader
	unknownKeySuggestMode common.UnknownKeySuggestionMode
	strict                bool
}

type UnknownKeySuggestionMode = common.UnknownKeySuggestionMode

const (
	Warn  UnknownKeySuggestionMode = common.UnknownKeySuggestionWarn
	Error UnknownKeySuggestionMode = common.UnknownKeySuggestionError
	Off   UnknownKeySuggestionMode = common.UnknownKeySuggestionOff
)

type fileSourceFactory func(path string, callerDir string, suggestionMode common.UnknownKeySuggestionMode) sourceLoader
type bytesSourceFactory func(data []byte, suggestionMode common.UnknownKeySuggestionMode) sourceLoader

func FromEnv() Option {
	return func(o *loadOptions) error {
		if o == nil {
			return errs.InvalidSchemaNilOptions
		}

		o.sources = append(o.sources, envsource.Load)
		return nil
	}
}

func FromDotEnvFile(path string) Option {
	return fromFile(path, errs.InvalidSchemaEmptyDotEnv, func(path string, callerDir string, _ common.UnknownKeySuggestionMode) sourceLoader {
		source := envsource.NewDotEnvFileSource(path, callerDir)
		return source.LoadFile
	})
}

func FromYAMLFile(path string) Option {
	return fromFile(path, errs.InvalidSchemaEmptyYAML, func(path string, callerDir string, suggestionMode common.UnknownKeySuggestionMode) sourceLoader {
		source := yamlsource.NewFileSource(path, callerDir, suggestionMode)
		return source.LoadFile
	})
}

func FromJSONFile(path string) Option {
	return fromFile(path, errs.InvalidSchemaEmptyJSON, func(path string, callerDir string, suggestionMode common.UnknownKeySuggestionMode) sourceLoader {
		source := jsonsource.NewFileSource(path, callerDir, suggestionMode)
		return source.LoadFile
	})
}

func FromTOMLFile(path string) Option {
	return fromFile(path, errs.InvalidSchemaEmptyTOML, func(path string, callerDir string, suggestionMode common.UnknownKeySuggestionMode) sourceLoader {
		source := tomlsource.NewFileSource(path, callerDir, suggestionMode)
		return source.LoadFile
	})
}

func FromYAMLBytes(data []byte) Option {
	return fromBytes(data, errs.InvalidSchemaEmptyYAMLBytes, func(data []byte, suggestionMode common.UnknownKeySuggestionMode) sourceLoader {
		source := yamlsource.NewByteSource(data, suggestionMode)
		return source.LoadBytes
	})
}

func FromJSONBytes(data []byte) Option {
	return fromBytes(data, errs.InvalidSchemaEmptyJSONBytes, func(data []byte, suggestionMode common.UnknownKeySuggestionMode) sourceLoader {
		source := jsonsource.NewByteSource(data, suggestionMode)
		return source.LoadBytes
	})
}

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
			load := factory(path, callerDir, mode)
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

func WithUnknownKeySuggestionMode(mode UnknownKeySuggestionMode) Option {
	return func(o *loadOptions) error {
		if o == nil {
			return errs.InvalidSchemaNilOptions
		}

		o.unknownKeySuggestMode = mode
		return nil
	}
}

func Strict() Option {
	return func(o *loadOptions) error {
		if o == nil {
			return errs.InvalidSchemaNilOptions
		}

		o.strict = true
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
