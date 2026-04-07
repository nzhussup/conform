package konform

import (
	internaldefaults "github.com/nzhussup/konform/internal/defaults"
	internalschema "github.com/nzhussup/konform/internal/schema"
	internalvalidate "github.com/nzhussup/konform/internal/validate"

	"github.com/nzhussup/konform/internal/errs"
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

	missing, err := internalvalidate.MissingRequired(sc)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		fieldErrors := make([]FieldError, 0, len(missing))
		for _, f := range missing {
			fieldErrors = append(fieldErrors, FieldError{
				Path: f.Path,
				Err:  errs.ValidationRequired,
			})
		}

		return &ValidationError{Fields: fieldErrors}
	}

	return nil
}
