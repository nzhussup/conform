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

func TestNonZero(t *testing.T) {
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
	makeFloatField := func(name string, value float64, rules map[string]string) schema.Field {
		v := value
		return schema.Field{
			GoName:      name,
			Path:        name,
			Validations: rules,
			Type:        reflect.TypeOf(0.0),
			Value:       reflect.ValueOf(&v).Elem(),
		}
	}
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

	tests := []struct {
		name        string
		field       schema.Field
		initial     []types.ValidationResult
		wantTotal   int
		wantErrType error
		wantLike    string
	}{
		{
			name:      "non-zero int passes",
			field:     makeIntField("Retries", 3, map[string]string{"nonzero": ""}),
			wantTotal: 0,
		},
		{
			name:      "non-zero float passes",
			field:     makeFloatField("Rate", 0.1, map[string]string{"nonzero": ""}),
			wantTotal: 0,
		},
		{
			name:        "zero int returns nonzero validation error",
			field:       makeIntField("Retries", 0, map[string]string{"nonzero": ""}),
			wantTotal:   1,
			wantErrType: errs.ValidationNonZero,
		},
		{
			name:        "zero float returns nonzero validation error",
			field:       makeFloatField("Rate", 0, map[string]string{"nonzero": ""}),
			wantTotal:   1,
			wantErrType: errs.ValidationNonZero,
		},
		{
			name:        "non-numeric field returns non-numeric validation error",
			field:       makeStringField("Name", "konform", map[string]string{"nonzero": ""}),
			wantTotal:   1,
			wantErrType: errs.ValidationNonNumeric,
			wantLike:    "nonzero validation supports only numeric values",
		},
		{
			name:  "appends after existing validations",
			field: makeIntField("Retries", 0, map[string]string{"nonzero": ""}),
			initial: []types.ValidationResult{
				{Err: errors.New("existing")},
			},
			wantTotal:   2,
			wantErrType: errs.ValidationNonZero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := append([]types.ValidationResult(nil), tt.initial...)

			validators.NonZero(tt.field, &results)

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
