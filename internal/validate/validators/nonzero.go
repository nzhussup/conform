package validators

import (
	"fmt"

	"github.com/nzhussup/konform/internal/errs"
	schematypes "github.com/nzhussup/konform/internal/schema/types"
	"github.com/nzhussup/konform/internal/validate/types"
)

const NonZeroRuleName = "nonzero"

func NonZero(f schematypes.Field, validations *[]types.ValidationResult) {
	if !isNumericKind(f.Type.Kind()) {
		*validations = append(*validations, types.ValidationResult{
			Field: f,
			Err:   fmt.Errorf("%w: %s validation supports only numeric values", errs.ValidationNonNumeric, NonZeroRuleName),
		})
		return
	}

	if schematypes.IsZeroValue(f.Value) {
		*validations = append(*validations, types.ValidationResult{
			Field: f,
			Err:   errs.ValidationNonZero,
		})
	}
}
