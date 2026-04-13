package schema

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/nzhussup/konform/internal/errs"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		name    string
		target  any
		errType error
	}{
		{
			name:    "nil interface",
			target:  nil,
			errType: errs.InvalidTarget,
		},
		{
			name:    "non-pointer target",
			target:  struct{}{},
			errType: errs.InvalidTarget,
		},
		{
			name:    "nil pointer target",
			target:  (*struct{})(nil),
			errType: errs.InvalidTarget,
		},
		{
			name:    "pointer to non-struct",
			target:  new(int),
			errType: errs.InvalidTarget,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := Build(tt.target)
			if s != nil {
				t.Fatalf("Build() expected nil schema on error, got %#v", s)
			}
			if !errors.Is(err, tt.errType) {
				t.Fatalf("Build() error = %v, want wrapped %v", err, tt.errType)
			}
		})
	}
}

func TestBuildCollectsExportedAndNestedFields(t *testing.T) {
	type nested struct {
		Flag   bool   `env:"FLAG" validate:"required"`
		hidden string `env:"HIDDEN"` //nolint:unused // intentionally unexported test field
	}

	type config struct {
		Name       string `key:"name" env:"NAME" default:"app" validate:"required"`
		Count      int
		Nested     nested `key:"nested"`
		unexported string `key:"skip"` //nolint:unused // intentionally unexported test field
	}

	s, err := Build(&config{})
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}
	if s == nil {
		t.Fatalf("Build() got nil schema")
	}

	if got, want := len(s.Fields), 4; got != want {
		t.Fatalf("Build() field count = %d, want %d", got, want)
	}

	f0 := s.Fields[0]
	if f0.Path != "Name" {
		t.Fatalf("field[0].Path = %q, want %q", f0.Path, "Name")
	}
	if f0.KeyName != "name" || f0.EnvName != "NAME" {
		t.Fatalf("field[0] tags not parsed correctly: key=%q env=%q", f0.KeyName, f0.EnvName)
	}
	if !f0.HasDefaultValue() || f0.DefaultValue != "app" {
		t.Fatalf("field[0] default parsing incorrect: has=%v value=%q", f0.HasDefaultValue(), f0.DefaultValue)
	}
	if !f0.HasValidation("required") {
		t.Fatalf("field[0] required validation parsing incorrect")
	}
	if f0.IsSecret {
		t.Fatalf("field[0].IsSecret = true, want false")
	}

	f1 := s.Fields[1]
	if f1.Path != "Count" {
		t.Fatalf("field[1].Path = %q, want %q", f1.Path, "Count")
	}

	f2 := s.Fields[2]
	if f2.Path != "Nested" {
		t.Fatalf("field[2].Path = %q, want %q", f2.Path, "Nested")
	}
	if f2.Type.Kind() != reflect.Struct {
		t.Fatalf("field[2].Type.Kind() = %v, want %v", f2.Type.Kind(), reflect.Struct)
	}

	f3 := s.Fields[3]
	if f3.Path != "Nested.Flag" {
		t.Fatalf("field[3].Path = %q, want %q", f3.Path, "Nested.Flag")
	}
	if f3.EnvName != "FLAG" || !f3.HasValidation("required") {
		t.Fatalf("field[3] tags not parsed correctly: env=%q required=%v", f3.EnvName, f3.HasValidation("required"))
	}
}

func TestBuildParsesSecretTag(t *testing.T) {
	type config struct {
		APIKey  string `secret:"true"`
		APIKey2 string `secret:"1"`
		Public  string `secret:"false"`
	}

	s, err := Build(&config{})
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	if got := s.Fields[0].IsSecret; !got {
		t.Fatalf("field[0].IsSecret = %v, want true", got)
	}
	if got := s.Fields[1].IsSecret; !got {
		t.Fatalf("field[1].IsSecret = %v, want true", got)
	}
	if got := s.Fields[2].IsSecret; got {
		t.Fatalf("field[2].IsSecret = %v, want false", got)
	}
}

func TestBuildReturnsInvalidSchemaForUnsupportedValidateRule(t *testing.T) {
	type config struct {
		Age int `validate:"unknown=10"`
	}

	_, err := Build(&config{})
	if err == nil {
		t.Fatalf("Build() error = nil, want invalid schema error")
	}
	if !errors.Is(err, errs.InvalidSchema) {
		t.Fatalf("Build() error = %v, want wrapped %v", err, errs.InvalidSchema)
	}
	if !strings.Contains(err.Error(), "unsupported validate rule \"unknown\"") {
		t.Fatalf("Build() error = %q, want unsupported unknown rule message", err.Error())
	}
}

func TestBuildAllowsCustomValidateRuleWithSupportChecker(t *testing.T) {
	type config struct {
		Age int `validate:"custom_rule=10"`
	}

	_, err := Build(&config{}, func(name string) bool {
		return name == "custom_rule"
	})
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}
}

func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		name  string
		value reflect.Value
		want  bool
	}{
		{name: "zero int", value: reflect.ValueOf(0), want: true},
		{name: "non-zero int", value: reflect.ValueOf(1), want: false},
		{name: "zero string", value: reflect.ValueOf(""), want: true},
		{name: "non-zero string", value: reflect.ValueOf("x"), want: false},
		{name: "nil pointer", value: reflect.ValueOf((*int)(nil)), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsZeroValue(tt.value); got != tt.want {
				t.Fatalf("IsZeroValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseValidateTag(t *testing.T) {
	t.Run("empty tag returns nil", func(t *testing.T) {
		got, err := parseValidateTag("Field", "", nil)
		if err != nil {
			t.Fatalf("parseValidateTag() error = %v, want nil", err)
		}
		if got != nil {
			t.Fatalf("parseValidateTag() = %#v, want nil", got)
		}
	})

	t.Run("skips blank parts and blank keys", func(t *testing.T) {
		got, err := parseValidateTag("Field", " , =x , required , min= 2 ", func(string) bool { return true })
		if err != nil {
			t.Fatalf("parseValidateTag() error = %v, want nil", err)
		}
		if got["required"] != "" || got["min"] != "2" || len(got) != 2 {
			t.Fatalf("parseValidateTag() = %#v, want required and min=2", got)
		}
	})

	t.Run("non-empty tag with no valid rules returns nil", func(t *testing.T) {
		got, err := parseValidateTag("Field", " , =x ,   ", func(string) bool { return true })
		if err != nil {
			t.Fatalf("parseValidateTag() error = %v, want nil", err)
		}
		if got != nil {
			t.Fatalf("parseValidateTag() = %#v, want nil", got)
		}
	})
}

func TestCollectFieldsNestedErrorPropagation(t *testing.T) {
	type nested struct {
		Port int `validate:"unknownrule=1"`
	}
	type cfg struct {
		Nested nested
	}

	v := reflect.ValueOf(&cfg{}).Elem()
	fields := make([]Field, 0)
	err := collectFields(v, v.Type(), "", &fields, func(name string) bool {
		return name != "unknownrule"
	})
	if err == nil {
		t.Fatalf("collectFields() error = nil, want invalid schema error")
	}
	if !errors.Is(err, errs.InvalidSchema) {
		t.Fatalf("collectFields() error = %v, want wrapped %v", err, errs.InvalidSchema)
	}
}
