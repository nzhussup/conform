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

func TestEmail(t *testing.T) {
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
	makeNonStringField := func(name string, value int, rules map[string]string) schema.Field {
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
			name:      "valid email passes",
			field:     makeStringField("Email", "user@example.com", map[string]string{"email": ""}),
			wantTotal: 0,
		},
		{
			name:      "valid email with plus and subdomain passes",
			field:     makeStringField("Email", "user+ops@mail.example.co.uk", map[string]string{"email": ""}),
			wantTotal: 0,
		},
		{
			name:        "missing at symbol returns validation email",
			field:       makeStringField("Email", "userexample.com", map[string]string{"email": ""}),
			wantTotal:   1,
			wantErrType: errs.ValidationEmail,
		},
		{
			name:        "missing domain returns validation email",
			field:       makeStringField("Email", "user@", map[string]string{"email": ""}),
			wantTotal:   1,
			wantErrType: errs.ValidationEmail,
		},
		{
			name:        "non-string values are rejected",
			field:       makeNonStringField("Retries", 3, map[string]string{"email": ""}),
			wantTotal:   1,
			wantErrType: errs.ValidationNonString,
			wantLike:    "email validation supports only string values",
		},
		{
			name:  "appends after existing validations",
			field: makeStringField("Email", "invalid", map[string]string{"email": ""}),
			initial: []types.ValidationResult{
				{Err: errors.New("existing")},
			},
			wantTotal:   2,
			wantErrType: errs.ValidationEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := append([]types.ValidationResult(nil), tt.initial...)

			validators.Email(tt.field, &results)

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
