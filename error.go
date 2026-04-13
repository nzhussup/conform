package konform

import (
	"fmt"
	"strings"

	"github.com/nzhussup/konform/internal/errs"
)

var (
	// ErrInvalidTarget indicates target passed to Load is invalid.
	ErrInvalidTarget = errs.InvalidTarget
	// ErrInvalidSchema indicates schema configuration or tags are invalid.
	ErrInvalidSchema = errs.InvalidSchema
	// ErrDecode indicates source decode or type conversion failure.
	ErrDecode = errs.Decode
	// ErrValidation indicates one or more validation rules failed.
	ErrValidation = errs.Validation
)

// FieldError represents an error for a specific configuration field path.
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

// ValidationError contains all field-level validation failures.
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
