package validators

import (
	"fmt"
	"net/mail"
	"reflect"

	"github.com/nzhussup/konform/internal/errs"
	schematypes "github.com/nzhussup/konform/internal/schema/types"
	"github.com/nzhussup/konform/internal/validate/types"
)

const EmailRuleName = "email"

func Email(f schematypes.Field, validations *[]types.ValidationResult) {
	if f.Type.Kind() != reflect.String {
		*validations = append(*validations, types.ValidationResult{
			Field: f,
			Err:   fmt.Errorf("%w: %s validation supports only string values", errs.ValidationNonString, EmailRuleName),
		})
		return
	}

	_, err := mail.ParseAddress(f.Value.String())
	if err != nil {
		*validations = append(*validations, types.ValidationResult{
			Field: f,
			Err:   errs.ValidationEmail,
		})
	}
}
