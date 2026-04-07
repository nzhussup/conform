package konform

import (
	"errors"
	"strings"
	"testing"

	"github.com/nzhussup/konform/internal/errs"
)

func TestFieldErrorError(t *testing.T) {
	tests := []struct {
		name string
		in   FieldError
		want string
	}{
		{
			name: "with path strips nested validation prefix",
			in: FieldError{
				Path: "AppDebug",
				Err:  errs.ValidationRequired,
			},
			want: "AppDebug: required",
		},
		{
			name: "without path returns raw error",
			in: FieldError{
				Path: "",
				Err:  errs.ValidationRequired,
			},
			want: errs.ValidationRequired.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.Error()
			if got != tt.want {
				t.Fatalf("FieldError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name        string
		in          *ValidationError
		wantLike    []string
		wantPrefixN int
	}{
		{
			name:        "nil validation error",
			in:          nil,
			wantLike:    []string{ErrValidation.Error()},
			wantPrefixN: 1,
		},
		{
			name: "empty fields",
			in:   &ValidationError{},
			wantLike: []string{
				ErrValidation.Error(),
			},
			wantPrefixN: 1,
		},
		{
			name: "field messages are clean",
			in: &ValidationError{
				Fields: []FieldError{
					{Path: "AppDebug", Err: errs.ValidationRequired},
				},
			},
			wantLike:    []string{"AppDebug: required"},
			wantPrefixN: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.in.Error()
			for _, part := range tt.wantLike {
				if !strings.Contains(msg, part) {
					t.Fatalf("ValidationError.Error() = %q, want to contain %q", msg, part)
				}
			}
			if got := strings.Count(msg, errs.Validation.Error()); got != tt.wantPrefixN {
				t.Fatalf("ValidationError.Error() validation prefix count = %d, want %d", got, tt.wantPrefixN)
			}
		})
	}
}

func TestValidationErrorUnwrap(t *testing.T) {
	tests := []struct {
		name string
		in   *ValidationError
		want error
	}{
		{name: "always unwraps to validation", in: &ValidationError{}, want: ErrValidation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.Unwrap(); !errors.Is(got, tt.want) {
				t.Fatalf("ValidationError.Unwrap() = %v, want wrapped %v", got, tt.want)
			}
		})
	}
}
