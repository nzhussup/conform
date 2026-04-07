package decode

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/nzhussup/conform/internal/errs"
	"github.com/nzhussup/conform/internal/schema"
)

func SetFieldValue(field schema.Field, raw string) error {
	if !field.Value.CanSet() {
		return errs.DecodeFieldCannotSet
	}

	switch field.Type.Kind() {
	case reflect.String:
		field.Value.SetString(raw)
		return nil

	case reflect.Int:
		v, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("%w: %q", errs.DecodeInvalidInt, raw)
		}
		field.Value.SetInt(int64(v))
		return nil

	case reflect.Bool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("%w: %q", errs.DecodeInvalidBool, raw)
		}
		field.Value.SetBool(v)
		return nil

	default:
		return fmt.Errorf("%w: %s", errs.DecodeUnsupported, field.Type)
	}
}
