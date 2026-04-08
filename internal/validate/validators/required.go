package validators

import (
	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
	"github.com/nzhussup/konform/internal/validate/model"
)

func Required(f schema.Field, validations *[]model.ValidationResult) {
	if !f.HasValidation("required") {
		return
	}
	if schema.IsZeroValue(f.Value) {
		*validations = append(*validations, model.ValidationResult{
			Field: f,
			Err:   errs.ValidationRequired,
		})
	}
}
