package schema

import (
	"fmt"
	"reflect"

	"github.com/nzhussup/conform/internal/errs"
)

type Field struct {
	GoName          string
	Path            string
	ConfName        string
	EnvName         string
	DefaultValue    string
	HasDefaultValue bool
	Required        bool
	Type            reflect.Type
	Value           reflect.Value
}

type Schema struct {
	Fields []Field
}

func Build(target any) (*Schema, error) {
	v := reflect.ValueOf(target)
	if !v.IsValid() || v.Kind() != reflect.Pointer || v.IsNil() {
		return nil, fmt.Errorf("%w: target must be a non-nil pointer to a struct", errs.InvalidTarget)
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: target must point to a struct", errs.InvalidTarget)
	}

	s := &Schema{}
	if err := collectFields(v, v.Type(), "", &s.Fields); err != nil {
		return nil, err
	}
	return s, nil
}

func IsZeroValue(v reflect.Value) bool {
	return v.IsZero()
}

func collectFields(v reflect.Value, t reflect.Type, parentPath string, fields *[]Field) error {
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
		hasDefaultValue := defaultValue != ""

		*fields = append(*fields, Field{
			GoName:          structField.Name,
			Path:            path,
			ConfName:        structField.Tag.Get("conf"),
			EnvName:         structField.Tag.Get("env"),
			DefaultValue:    defaultValue,
			HasDefaultValue: hasDefaultValue,
			Required:        structField.Tag.Get("required") == "true",
			Type:            structField.Type,
			Value:           fieldValue,
		})

		if structField.Type.Kind() == reflect.Struct {
			if err := collectFields(fieldValue, structField.Type, path, fields); err != nil {
				return err
			}
		}
	}
	return nil
}
