package schema

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nzhussup/konform/internal/errs"
	schematypes "github.com/nzhussup/konform/internal/schema/types"
	"github.com/nzhussup/konform/internal/validate/rules"
)

type Field = schematypes.Field
type Schema = schematypes.Schema

func Build(target any, isSupportedRule ...func(string) bool) (*Schema, error) {
	supportedRule := rules.IsSupported
	if len(isSupportedRule) > 0 && isSupportedRule[0] != nil {
		supportedRule = isSupportedRule[0]
	}

	v := reflect.ValueOf(target)
	if !v.IsValid() || v.Kind() != reflect.Pointer || v.IsNil() {
		return nil, fmt.Errorf("%w: target must be a non-nil pointer to a struct", errs.InvalidTarget)
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: target must point to a struct", errs.InvalidTarget)
	}

	s := &Schema{}
	if err := collectFields(v, v.Type(), "", &s.Fields, supportedRule); err != nil {
		return nil, err
	}
	return s, nil
}

func IsZeroValue(v reflect.Value) bool {
	return schematypes.IsZeroValue(v)
}

func parseValidateTag(path string, tag string, isSupportedRule func(string) bool) (map[string]string, error) {
	if tag == "" {
		return nil, nil
	}

	parsedRules := map[string]string{}
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		k, v, hasValue := strings.Cut(part, "=")
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if isSupportedRule != nil && !isSupportedRule(k) {
			return nil, fmt.Errorf("%w: unsupported validate rule %q for field %q", errs.InvalidSchema, k, path)
		}
		if hasValue {
			parsedRules[k] = strings.TrimSpace(v)
			continue
		}
		parsedRules[k] = ""
	}

	if len(parsedRules) == 0 {
		return nil, nil
	}
	return parsedRules, nil
}

func collectFields(v reflect.Value, t reflect.Type, parentPath string, fields *[]Field, isSupportedRule func(string) bool) error {
	for i := 0; i < t.NumField(); i++ {
		structField := t.Field(i)
		fieldValue := v.Field(i)

		if structField.PkgPath != "" {
			continue
		}

		path := structField.Name
		if parentPath != "" {
			path = parentPath + "." + structField.Name
		}

		defaultValue := structField.Tag.Get("default")
		validations, err := parseValidateTag(path, structField.Tag.Get("validate"), isSupportedRule)
		if err != nil {
			return err
		}

		*fields = append(*fields, Field{
			GoName:       structField.Name,
			Path:         path,
			KeyName:      structField.Tag.Get("key"),
			EnvName:      structField.Tag.Get("env"),
			DefaultValue: defaultValue,
			IsSecret:     parseSecretTag(structField.Tag.Get("secret")),
			Validations:  validations,
			Type:         structField.Type,
			Value:        fieldValue,
		})

		if structField.Type.Kind() == reflect.Struct {
			if err := collectFields(fieldValue, structField.Type, path, fields, isSupportedRule); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseSecretTag(tag string) bool {
	switch strings.ToLower(strings.TrimSpace(tag)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
