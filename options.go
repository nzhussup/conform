package conform

import (
	"github.com/nzhussup/conform/internal/errs"
	internalschema "github.com/nzhussup/conform/internal/schema"
	envsource "github.com/nzhussup/conform/internal/source/env"

	yamlsource "github.com/nzhussup/conform/internal/source/yaml"
)

type Option func(*loadOptions) error

type sourceLoader func(*internalschema.Schema) error

type loadOptions struct {
	sources []sourceLoader
}

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
	return func(o *loadOptions) error {
		if o == nil {
			return errs.InvalidSchemaNilOptions
		}

		if path == "" {
			return errs.InvalidSchemaEmptyYAML
		}

		source := yamlsource.NewFileSource(path)
		o.sources = append(o.sources, source.Load)
		return nil
	}
}
