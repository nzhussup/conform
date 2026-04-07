package konform

import (
	"fmt"
	"strings"

	"github.com/nzhussup/konform/internal/errs"
)

var (
	ErrInvalidTarget = errs.InvalidTarget
	ErrInvalidSchema = errs.InvalidSchema
	ErrDecode        = errs.Decode
	ErrValidation    = errs.Validation
)

type FieldError struct {
	Path string
	Err  error
}

func (e FieldError) Error() string {
	if e.Path == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %s", e.Path, errs.StripValidationPrefix(e.Err))
}

type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	if e == nil || len(e.Fields) == 0 {
		return ErrValidation.Error()
	}

	var b strings.Builder
	b.WriteString("konform: validation failed:")
	for _, field := range e.Fields {
		b.WriteString("\n  - ")
		b.WriteString(field.Error())
	}
	return b.String()
}

func (e *ValidationError) Unwrap() error {
	return ErrValidation
}
