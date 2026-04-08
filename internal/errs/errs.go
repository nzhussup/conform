package errs

import (
	"errors"
	"fmt"
	"strings"
)

var (
	InvalidTarget = errors.New("konform: invalid target")
	InvalidSchema = errors.New("konform: invalid schema")
	Decode        = errors.New("konform: decode error")
	Validation    = errors.New("konform: validation failed")

	InvalidSchemaNil        = fmt.Errorf("%w: nil schema", InvalidSchema)
	InvalidSchemaNilOptions = fmt.Errorf("%w: nil load options", InvalidSchema)
	InvalidSchemaEmptyYAML  = fmt.Errorf("%w: yaml path must not be empty", InvalidSchema)
	InvalidSchemaEmptyJSON  = fmt.Errorf("%w: json path must not be empty", InvalidSchema)
	InvalidSchemaEmptyTOML  = fmt.Errorf("%w: toml path must not be empty", InvalidSchema)
	ValidationRequired      = fmt.Errorf("%w: required", Validation)

	DecodeFieldCannotSet  = fmt.Errorf("%w: field cannot be set", Decode)
	DecodeInvalidInt      = fmt.Errorf("%w: invalid int value", Decode)
	DecodeInvalidFloat    = fmt.Errorf("%w: invalid float value", Decode)
	DecodeInvalidBool     = fmt.Errorf("%w: invalid bool value", Decode)
	DecodeInvalidDuration = fmt.Errorf("%w: invalid duration value", Decode)
	DecodeTypeMismatch    = fmt.Errorf("%w: type mismatch", Decode)
	DecodeUnsupported     = fmt.Errorf("%w: unsupported field type", Decode)
	DecodeSourceRead      = fmt.Errorf("%w: failed to read source file", Decode)
	DecodeSourceParse     = fmt.Errorf("%w: failed to parse source file", Decode)
	DecodeSourceField     = fmt.Errorf("%w: failed to decode source field", Decode)
)

type decodeContextError struct {
	kind    error
	context string
	cause   error
}

func (e *decodeContextError) Error() string {
	base := Decode.Error()
	if e.context != "" {
		base += ": " + e.context
	}
	if e.cause != nil {
		base += ": " + StripDecodePrefix(e.cause)
	}
	return base
}

func (e *decodeContextError) Unwrap() error {
	return e.cause
}

func (e *decodeContextError) Is(target error) bool {
	if target == Decode {
		return true
	}
	return e.kind != nil && target == e.kind
}

func WrapDecode(kind error, context string, cause error) error {
	return &decodeContextError{
		kind:    kind,
		context: context,
		cause:   cause,
	}
}

func StripDecodePrefix(err error) string {
	return StripDomainPrefix(err, Decode)
}

func StripValidationPrefix(err error) string {
	return StripDomainPrefix(err, Validation)
}

func StripDomainPrefix(err error, domain error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	prefix := domain.Error() + ": "
	for strings.HasPrefix(msg, prefix) {
		msg = strings.TrimPrefix(msg, prefix)
	}
	return msg
}
