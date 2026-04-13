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

func TestUrl(t *testing.T) {
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

	makeNonStringField := func(name string, rules map[string]string) schema.Field {
		v := 8080
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
			name:      "missing url rule is ignored",
			field:     makeStringField("URL", "https://example.com", nil),
			wantTotal: 0,
		},
		{
			name:      "http url with host passes",
			field:     makeStringField("URL", "http://example.com/api", map[string]string{"url": ""}),
			wantTotal: 0,
		},
		{
			name:      "https url with port passes",
			field:     makeStringField("URL", "https://example.com:8443/api", map[string]string{"url": ""}),
			wantTotal: 0,
		},
		{
			name:      "url is trimmed before validation",
			field:     makeStringField("URL", "  https://example.com/path  ", map[string]string{"url": ""}),
			wantTotal: 0,
		},
		{
			name:        "ws scheme is rejected",
			field:       makeStringField("URL", "ws://example.com/socket", map[string]string{"url": ""}),
			wantTotal:   1,
			wantErrType: errs.ValidationURL,
			wantLike:    "allowed schemes [http https]",
		},
		{
			name:        "missing host is rejected",
			field:       makeStringField("URL", "http://", map[string]string{"url": ""}),
			wantTotal:   1,
			wantErrType: errs.ValidationURL,
			wantLike:    "allowed schemes [http https]",
		},
		{
			name:        "http scheme with empty host and path is rejected",
			field:       makeStringField("URL", "http:///only-path", map[string]string{"url": ""}),
			wantTotal:   1,
			wantErrType: errs.ValidationURL,
			wantLike:    "allowed schemes [http https]",
		},
		{
			name:        "invalid url syntax is rejected",
			field:       makeStringField("URL", "htp:/invalid-url", map[string]string{"url": ""}),
			wantTotal:   1,
			wantErrType: errs.ValidationURL,
			wantLike:    "invalid URL format",
		},
		{
			name:        "non-string values are rejected",
			field:       makeNonStringField("Port", map[string]string{"url": ""}),
			wantTotal:   1,
			wantErrType: errs.ValidationNonString,
			wantLike:    "url validation supports only string values",
		},
		{
			name:  "appends after existing validations",
			field: makeStringField("URL", "ws://example.com/socket", map[string]string{"url": ""}),
			initial: []types.ValidationResult{
				{Err: errors.New("existing")},
			},
			wantTotal:   2,
			wantErrType: errs.ValidationURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := append([]types.ValidationResult(nil), tt.initial...)

			validators.URL(tt.field, &results)

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
