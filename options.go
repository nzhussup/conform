package konform

import (
	"path/filepath"
	"runtime"

	"github.com/nzhussup/konform/internal/errs"
	internalschema "github.com/nzhussup/konform/internal/schema"
	envsource "github.com/nzhussup/konform/internal/source/env"

	"github.com/nzhussup/konform/internal/source/common"
	jsonsource "github.com/nzhussup/konform/internal/source/json"
	tomlsource "github.com/nzhussup/konform/internal/source/toml"
	yamlsource "github.com/nzhussup/konform/internal/source/yaml"
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

func FromEnv() Option {
	return func(o *loadOptions) error {
		if o == nil {
			return errs.InvalidSchemaNilOptions
		}

		o.sources = append(o.sources, envsource.Load)
		return nil
	}
}

func FromYAMLFile(path string) Option {
	return fromFile(path, errs.InvalidSchemaEmptyYAML, func(path string, callerDir string, suggestionMode common.UnknownKeySuggestionMode) sourceLoader {
		source := yamlsource.NewFileSource(path, callerDir, suggestionMode)
		return source.Load
	})
}

func FromJSONFile(path string) Option {
	return fromFile(path, errs.InvalidSchemaEmptyJSON, func(path string, callerDir string, suggestionMode common.UnknownKeySuggestionMode) sourceLoader {
		source := jsonsource.NewFileSource(path, callerDir, suggestionMode)
		return source.Load
	})
}

func FromTOMLFile(path string) Option {
	return fromFile(path, errs.InvalidSchemaEmptyTOML, func(path string, callerDir string, suggestionMode common.UnknownKeySuggestionMode) sourceLoader {
		source := tomlsource.NewFileSource(path, callerDir, suggestionMode)
		return source.Load
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
