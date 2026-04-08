package validators

import (
	"errors"
	"reflect"
	"testing"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
	"github.com/nzhussup/konform/internal/validate/model"
)

func TestRequired(t *testing.T) {
	makeStringField := func(name string, rules map[string]string, target *string) schema.Field {
		return schema.Field{
			GoName:      name,
			Path:        name,
			Validations: rules,
			Type:        reflect.TypeOf(""),
			Value:       reflect.ValueOf(target).Elem(),
		}
	}
	makeIntField := func(name string, rules map[string]string, target *int) schema.Field {
		return schema.Field{
			GoName:      name,
			Path:        name,
			Validations: rules,
			Type:        reflect.TypeOf(0),
			Value:       reflect.ValueOf(target).Elem(),
		}
	}

	tests := []struct {
		name        string
		field       schema.Field
		initial     []model.ValidationResult
		wantAdded   bool
		wantTotal   int
		wantErrType error
	}{
		{
			name:      "optional field is ignored even when zero",
			field:     makeStringField("Name", nil, new(string)),
			initial:   nil,
			wantAdded: false,
			wantTotal: 0,
		},
		{
			name: "required field with non-zero value has no validation error",
			field: func() schema.Field {
				v := "konform"
				return makeStringField("Name", map[string]string{"required": ""}, &v)
			}(),
			initial:   nil,
			wantAdded: false,
			wantTotal: 0,
		},
		{
			name:        "required zero string adds required validation result",
			field:       makeStringField("Name", map[string]string{"required": ""}, new(string)),
			initial:     nil,
			wantAdded:   true,
			wantTotal:   1,
			wantErrType: errs.ValidationRequired,
		},
		{
			name:        "required zero int adds required validation result",
			field:       makeIntField("Port", map[string]string{"required": ""}, new(int)),
			initial:     nil,
			wantAdded:   true,
			wantTotal:   1,
			wantErrType: errs.ValidationRequired,
		},
		{
			name:  "required failure appends after existing validations",
			field: makeStringField("Name", map[string]string{"required": ""}, new(string)),
			initial: []model.ValidationResult{
				{Err: errors.New("existing")},
			},
			wantAdded:   true,
			wantTotal:   2,
			wantErrType: errs.ValidationRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := append([]model.ValidationResult(nil), tt.initial...)

			Required(tt.field, &results)

			if got := len(results); got != tt.wantTotal {
				t.Fatalf("len(results) = %d, want %d", got, tt.wantTotal)
			}

			if !tt.wantAdded {
				return
			}

			got := results[len(results)-1]
			if got.Field.GoName != tt.field.GoName {
				t.Fatalf("result field GoName = %q, want %q", got.Field.GoName, tt.field.GoName)
			}
			if got.Field.Path != tt.field.Path {
				t.Fatalf("result field Path = %q, want %q", got.Field.Path, tt.field.Path)
			}
			if !errors.Is(got.Err, tt.wantErrType) {
				t.Fatalf("result error = %v, want wrapped %v", got.Err, tt.wantErrType)
			}
		})
	}
}
