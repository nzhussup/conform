package errs

import (
	"errors"
	"fmt"
)

var (
	InvalidTarget = errors.New("conform: invalid target")
	InvalidSchema = errors.New("conform: invalid schema")
	Decode        = errors.New("conform: decode error")
	Validation    = errors.New("conform: validation failed")

	InvalidSchemaNil        = fmt.Errorf("%w: nil schema", InvalidSchema)
	InvalidSchemaNilOptions = fmt.Errorf("%w: nil load options", InvalidSchema)
	InvalidSchemaEmptyYAML  = fmt.Errorf("%w: yaml path must not be empty", InvalidSchema)
	ValidationRequired      = fmt.Errorf("%w: required", Validation)

	DecodeFieldCannotSet = fmt.Errorf("%w: field cannot be set", Decode)
	DecodeInvalidInt     = fmt.Errorf("%w: invalid int value", Decode)
	DecodeInvalidBool    = fmt.Errorf("%w: invalid bool value", Decode)
	DecodeUnsupported    = fmt.Errorf("%w: unsupported field type", Decode)
)
