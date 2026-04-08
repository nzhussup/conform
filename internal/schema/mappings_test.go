package schema

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/nzhussup/konform/internal/errs"
)

func TestValidateStrictMappings(t *testing.T) {
	t.Run("nil schema", func(t *testing.T) {
		err := ValidateStrictMappings(nil)
		if !errors.Is(err, errs.InvalidSchemaNil) {
			t.Fatalf("ValidateStrictMappings() error = %v, want %v", err, errs.InvalidSchemaNil)
		}
	})

	t.Run("duplicate key mapping", func(t *testing.T) {
		var a string
		var b string
		sc := &Schema{
			Fields: []Field{
				{
					Path:    "A",
					KeyName: "app.name",
					Type:    reflect.TypeOf(""),
					Value:   reflect.ValueOf(&a).Elem(),
				},
				{
					Path:    "B",
					KeyName: "app.name",
					Type:    reflect.TypeOf(""),
					Value:   reflect.ValueOf(&b).Elem(),
				},
			},
		}

		err := ValidateStrictMappings(sc)
		if err == nil {
			t.Fatalf("ValidateStrictMappings() error = nil, want invalid schema error")
		}
		if !errors.Is(err, errs.InvalidSchema) {
			t.Fatalf("ValidateStrictMappings() error = %v, want wrapped %v", err, errs.InvalidSchema)
		}
		if !strings.Contains(err.Error(), `conflicting key mapping "app.name"`) {
			t.Fatalf("ValidateStrictMappings() error = %q, want key conflict details", err.Error())
		}
	})

	t.Run("duplicate env mapping", func(t *testing.T) {
		var a string
		var b string
		sc := &Schema{
			Fields: []Field{
				{
					Path:    "A",
					EnvName: "APP_NAME",
					Type:    reflect.TypeOf(""),
					Value:   reflect.ValueOf(&a).Elem(),
				},
				{
					Path:    "B",
					EnvName: "APP_NAME",
					Type:    reflect.TypeOf(""),
					Value:   reflect.ValueOf(&b).Elem(),
				},
			},
		}

		err := ValidateStrictMappings(sc)
		if err == nil {
			t.Fatalf("ValidateStrictMappings() error = nil, want invalid schema error")
		}
		if !errors.Is(err, errs.InvalidSchema) {
			t.Fatalf("ValidateStrictMappings() error = %v, want wrapped %v", err, errs.InvalidSchema)
		}
		if !strings.Contains(err.Error(), `conflicting env mapping "APP_NAME"`) {
			t.Fatalf("ValidateStrictMappings() error = %q, want env conflict details", err.Error())
		}
	})

	t.Run("no conflicts", func(t *testing.T) {
		var a string
		var b string
		sc := &Schema{
			Fields: []Field{
				{
					Path:    "A",
					KeyName: "app.name",
					EnvName: "APP_NAME",
					Type:    reflect.TypeOf(""),
					Value:   reflect.ValueOf(&a).Elem(),
				},
				{
					Path:    "B",
					KeyName: "app.mode",
					EnvName: "APP_MODE",
					Type:    reflect.TypeOf(""),
					Value:   reflect.ValueOf(&b).Elem(),
				},
			},
		}

		if err := ValidateStrictMappings(sc); err != nil {
			t.Fatalf("ValidateStrictMappings() error = %v, want nil", err)
		}
	})
}
