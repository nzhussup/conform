package conform

import (
	"github.com/nzhussup/conform/internal/errs"
	internalschema "github.com/nzhussup/conform/internal/schema"
)

func validateRequired(sc *internalschema.Schema) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	var missing []FieldError

	for _, f := range sc.Fields {
		if !f.Required {
			continue
		}

		if internalschema.IsZeroValue(f.Value) {
			missing = append(missing, FieldError{
				Path: f.Path,
				Err:  errs.ValidationRequired,
			})
		}
	}

	if len(missing) > 0 {
		return &ValidationError{
			Fields: missing,
		}
	}

	return nil
}
