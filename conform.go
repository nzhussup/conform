package conform

import (
	internaldefaults "github.com/nzhussup/conform/internal/defaults"
	internalschema "github.com/nzhussup/conform/internal/schema"
)

func Load(target any, opts ...Option) error {
	loadOpts := loadOptions{}

	for _, opt := range opts {
		if err := opt(&loadOpts); err != nil {
			return err
		}
	}

	sc, err := internalschema.Build(target)
	if err != nil {
		return err
	}

	if err := internaldefaults.Apply(sc); err != nil {
		return err
	}

	for _, src := range loadOpts.sources {
		if err := src(sc); err != nil {
			return err
		}
	}

	if err := validateRequired(sc); err != nil {
		return err
	}

	return nil
}
