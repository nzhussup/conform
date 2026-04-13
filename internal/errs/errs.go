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

	InvalidSchemaNil            = fmt.Errorf("%w: nil schema", InvalidSchema)
	InvalidSchemaNilOptions     = fmt.Errorf("%w: nil load options", InvalidSchema)
	InvalidSchemaEmptyYAML      = fmt.Errorf("%w: yaml path must not be empty", InvalidSchema)
	InvalidSchemaEmptyJSON      = fmt.Errorf("%w: json path must not be empty", InvalidSchema)
	InvalidSchemaEmptyTOML      = fmt.Errorf("%w: toml path must not be empty", InvalidSchema)
	InvalidSchemaEmptyDotEnv    = fmt.Errorf("%w: .env path must not be empty", InvalidSchema)
	InvalidSchemaEmptyYAMLBytes = fmt.Errorf("%w: yaml bytes must not be empty", InvalidSchema)
	InvalidSchemaEmptyJSONBytes = fmt.Errorf("%w: json bytes must not be empty", InvalidSchema)
	InvalidSchemaEmptyTOMLBytes = fmt.Errorf("%w: toml bytes must not be empty", InvalidSchema)

	ValidationRequired   = fmt.Errorf("%w: required", Validation)
	ValidationMin        = fmt.Errorf("%w: minimum value not met", Validation)
	ValidationMax        = fmt.Errorf("%w: maximum value exceeded", Validation)
	ValidationNonNumeric = fmt.Errorf("%w: non-numeric value", Validation)
	ValidationNonString  = fmt.Errorf("%w: non-string value", Validation)
	ValidationLen        = fmt.Errorf("%w: length mismatch", Validation)
	ValidationMinLen     = fmt.Errorf("%w: minimum length not met", Validation)
	ValidationMaxLen     = fmt.Errorf("%w: maximum length exceeded", Validation)
	ValidationRegex      = fmt.Errorf("%w: regex mismatch", Validation)
	ValidationOneOf      = fmt.Errorf("%w: none of the allowed values matched", Validation)
	ValidationNonZero    = fmt.Errorf("%w: zero value not allowed", Validation)
	ValidationURL        = fmt.Errorf("%w: invalid URL format", Validation)
	ValidationEmail      = fmt.Errorf("%w: invalid email format", Validation)

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
