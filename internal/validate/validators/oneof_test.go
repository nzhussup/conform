package validators_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
	"github.com/nzhussup/konform/internal/validate/types"
	"github.com/nzhussup/konform/internal/validate/validators"
)

func TestOneOf(t *testing.T) {
	makeStringField := func(name, value string, rules map[string]string) schema.Field {
		v := value
		return schema.Field{
			GoName:      name,
			Path:        name,
			Validations: rules,
			Type:        reflect.TypeOf(""),
			Value:       reflect.ValueOf(&v).Elem(),
		}
	}
	makeIntField := func(name string, value int, rules map[string]string) schema.Field {
		v := value
		return schema.Field{
			GoName:      name,
			Path:        name,
			Validations: rules,
			Type:        reflect.TypeOf(0),
			Value:       reflect.ValueOf(&v).Elem(),
		}
	}

	tests := []struct {
		name        string
		field       schema.Field
		initial     []types.ValidationResult
		wantTotal   int
		wantErrType error
		wantLike    string
	}{
		{
			name:      "missing oneof rule is ignored",
			field:     makeStringField("Mode", "dev", nil),
			wantTotal: 0,
		},
		{
			name:      "exact option match passes",
			field:     makeStringField("Mode", "prod", map[string]string{"oneof": "dev|prod|stage"}),
			wantTotal: 0,
		},
		{
			name:      "whitespace around options is trimmed",
			field:     makeStringField("Mode", "prod", map[string]string{"oneof": "dev | prod | stage"}),
			wantTotal: 0,
		},
		{
			name:      "numeric field value can match oneof values",
			field:     makeIntField("Port", 5432, map[string]string{"oneof": "3306|5432|27017"}),
			wantTotal: 0,
		},
		{
			name:      "empty oneof arg allows empty string value",
			field:     makeStringField("Label", "", map[string]string{"oneof": ""}),
			wantTotal: 0,
		},
		{
			name:        "empty oneof arg rejects non-empty value",
			field:       makeStringField("Label", "non-empty", map[string]string{"oneof": ""}),
			wantTotal:   1,
			wantErrType: errs.ValidationOneOf,
			wantLike:    `expected one of [] but got "non-empty"`,
		},
		{
			name:        "value outside allowed set returns validation oneof",
			field:       makeStringField("Mode", "qa", map[string]string{"oneof": "dev|prod|stage"}),
			wantTotal:   1,
			wantErrType: errs.ValidationOneOf,
			wantLike:    `expected one of [dev prod stage] but got "qa"`,
		},
		{
			name:  "appends after existing validations",
			field: makeStringField("Mode", "qa", map[string]string{"oneof": "dev|prod|stage"}),
			initial: []types.ValidationResult{
				{Err: errors.New("existing")},
			},
			wantTotal:   2,
			wantErrType: errs.ValidationOneOf,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := append([]types.ValidationResult(nil), tt.initial...)

			validators.OneOf(tt.field, &results)

			if got := len(results); got != tt.wantTotal {
				t.Fatalf("len(results) = %d, want %d", got, tt.wantTotal)
			}
			if tt.wantErrType == nil {
				return
			}

			got := results[len(results)-1]
			if got.Field.Path != tt.field.Path {
				t.Fatalf("result field Path = %q, want %q", got.Field.Path, tt.field.Path)
			}
			if !errors.Is(got.Err, tt.wantErrType) {
				t.Fatalf("result error = %v, want wrapped %v", got.Err, tt.wantErrType)
			}
			if tt.wantLike != "" && !strings.Contains(got.Err.Error(), tt.wantLike) {
				t.Fatalf("result error = %q, want to contain %q", got.Err.Error(), tt.wantLike)
			}
		})
	}
}
