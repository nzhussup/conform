package types

import (
	"fmt"
	"reflect"
)

type Field struct {
	GoName       string
	Path         string
	KeyName      string
	EnvName      string
	DefaultValue string
	IsSecret     bool
	Validations  map[string]string
	Type         reflect.Type
	Value        reflect.Value
	Source       string
}

func (f Field) HasDefaultValue() bool {
	return f.DefaultValue != ""
}

func (f Field) HasValidation(name string) bool {
	if len(f.Validations) == 0 {
		return false
	}
	_, ok := f.Validations[name]
	return ok
}

func (f Field) ValidationArg(name string) (string, bool) {
	if len(f.Validations) == 0 {
		return "", false
	}
	v, ok := f.Validations[name]
	return v, ok
}

func (f Field) ValueAsString() string {
	return fmt.Sprintf("%v", f.Value.Interface())
}

type Schema struct {
	Fields []Field
}

func IsZeroValue(v reflect.Value) bool {
	return v.IsZero()
}
