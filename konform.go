package konform

import (
	"maps"

	internaldefaults "github.com/nzhussup/konform/internal/defaults"
	internalschema "github.com/nzhussup/konform/internal/schema"
	internalvalidate "github.com/nzhussup/konform/internal/validate"
	internalvalidaterules "github.com/nzhussup/konform/internal/validate/rules"
	internalvalidatetypes "github.com/nzhussup/konform/internal/validate/types"
)

// Load builds a schema from target and applies defaults, configured sources,
// and validations in order.
//
// target must be a non-nil pointer to a struct.
func Load(target any, opts ...Option) (*LoadReport, error) {
	loadOpts := loadOptions{}

	for _, opt := range opts {
		if err := opt(&loadOpts); err != nil {
			return nil, err
		}
	}

	sc, err := internalschema.Build(target, makeRuleSupportChecker(loadOpts.customValidators))
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

	validatorRegistry := buildValidatorRegistry(loadOpts.customValidators)

	validations, err := internalvalidate.Validate(sc, validatorRegistry)
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

func makeRuleSupportChecker(custom map[string]CustomValidatorFunc) func(string) bool {
	return func(name string) bool {
		if _, ok := custom[name]; ok {
			return true
		}
		return internalvalidaterules.IsSupported(name)
	}
}

func buildValidatorRegistry(custom map[string]CustomValidatorFunc) map[string]internalvalidatetypes.ValidationFunc {
	registry := make(map[string]internalvalidatetypes.ValidationFunc, len(internalvalidaterules.Registry)+len(custom))
	maps.Copy(registry, internalvalidaterules.Registry)
	for name, validator := range custom {
		registry[name] = wrapCustomValidator(name, validator)
	}
	return registry
}

func wrapCustomValidator(ruleName string, validator CustomValidatorFunc) internalvalidatetypes.ValidationFunc {
	return func(field internalschema.Field, results *[]internalvalidatetypes.ValidationResult) {
		ruleValue := field.Validations[ruleName]
		if err := validator(field.Value.Interface(), ruleValue); err != nil {
			*results = append(*results, internalvalidatetypes.ValidationResult{
				Field: field,
				Err:   err,
			})
		}
	}
}
