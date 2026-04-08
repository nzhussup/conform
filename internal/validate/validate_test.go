package validate

import (
	"errors"
	"reflect"
	"testing"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

func TestValidate(t *testing.T) {
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
		name          string
		schemaBuilder func() *schema.Schema
		wantErrType   error
		wantCount     int
		wantFieldPath []string
	}{
		{
			name:          "nil schema returns invalid schema error",
			schemaBuilder: func() *schema.Schema { return nil },
			wantErrType:   errs.InvalidSchemaNil,
			wantCount:     0,
		},
		{
			name: "empty schema returns no validations",
			schemaBuilder: func() *schema.Schema {
				return &schema.Schema{}
			},
			wantCount: 0,
		},
		{
			name: "all optional fields return no validations",
			schemaBuilder: func() *schema.Schema {
				return &schema.Schema{
					Fields: []schema.Field{
						makeStringField("Name", nil, new(string)),
						makeIntField("Port", nil, new(int)),
					},
				}
			},
			wantCount: 0,
		},
		{
			name: "required zero fields are collected in order",
			schemaBuilder: func() *schema.Schema {
				return &schema.Schema{
					Fields: []schema.Field{
						makeStringField("Name", map[string]string{"required": ""}, new(string)),
						makeIntField("Port", map[string]string{"required": ""}, new(int)),
					},
				}
			},
			wantCount:     2,
			wantFieldPath: []string{"Name", "Port"},
		},
		{
			name: "mixed required and non-zero required fields",
			schemaBuilder: func() *schema.Schema {
				nonZeroName := "configured"
				return &schema.Schema{
					Fields: []schema.Field{
						makeStringField("Name", map[string]string{"required": ""}, &nonZeroName),
						makeIntField("Port", map[string]string{"required": ""}, new(int)),
						makeStringField("Env", nil, new(string)),
					},
				}
			},
			wantCount:     1,
			wantFieldPath: []string{"Port"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := tt.schemaBuilder()
			results, err := Validate(sc)

			if tt.wantErrType != nil {
				if err == nil {
					t.Fatalf("Validate() error = nil, want %v", tt.wantErrType)
				}
				if !errors.Is(err, tt.wantErrType) {
					t.Fatalf("Validate() error = %v, want wrapped %v", err, tt.wantErrType)
				}
				if len(results) != tt.wantCount {
					t.Fatalf("len(results) = %d, want %d", len(results), tt.wantCount)
				}
				return
			}

			if err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
			if len(results) != tt.wantCount {
				t.Fatalf("len(results) = %d, want %d", len(results), tt.wantCount)
			}
			for i, wantPath := range tt.wantFieldPath {
				if results[i].Field.Path != wantPath {
					t.Fatalf("results[%d].Field.Path = %q, want %q", i, results[i].Field.Path, wantPath)
				}
				if !errors.Is(results[i].Err, errs.ValidationRequired) {
					t.Fatalf("results[%d].Err = %v, want wrapped %v", i, results[i].Err, errs.ValidationRequired)
				}
			}
		})
	}
}
