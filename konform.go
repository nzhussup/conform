package konform

import (
	internaldefaults "github.com/nzhussup/konform/internal/defaults"
	internalschema "github.com/nzhussup/konform/internal/schema"
	internalvalidate "github.com/nzhussup/konform/internal/validate"
)

func Load(target any, opts ...Option) (*LoadReport, error) {
	loadOpts := loadOptions{}

	for _, opt := range opts {
		if err := opt(&loadOpts); err != nil {
			return nil, err
		}
	}

	sc, err := internalschema.Build(target)
	if err != nil {
		return nil, err
	}
	if loadOpts.strict {
		if err := internalschema.ValidateStrictMappings(sc); err != nil {
			return nil, err
		}
	}

	if err := internaldefaults.Apply(sc); err != nil {
		return buildReport(sc), err
	}

	for _, src := range loadOpts.sources {
		if err := src(sc); err != nil {
			return buildReport(sc), err
		}
	}

	validations, err := internalvalidate.Validate(sc)
	if err != nil {
		return buildReport(sc), err
	}
	if len(validations) > 0 {
		fieldErrors := make([]FieldError, 0, len(validations))
		for _, validation := range validations {
			fieldErrors = append(fieldErrors, FieldError{
				Path: validation.Field.Path,
				Err:  validation.Err,
			})
		}

		return buildReport(sc), &ValidationError{Fields: fieldErrors}
	}

	return buildReport(sc), nil
}
