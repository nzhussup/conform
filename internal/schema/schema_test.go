package schema

import (
	"errors"
	"reflect"
	"testing"

	"github.com/nzhussup/conform/internal/errs"
)

func TestBuild(t *testing.T) {
	type nested struct {
		Flag   bool   `env:"FLAG" required:"true"`
		hidden string `env:"HIDDEN"`
	}

	type config struct {
		Name       string `conf:"name" env:"NAME" default:"app" required:"true"`
		Count      int
		Nested     nested `conf:"nested"`
		unexported string `conf:"skip"`
	}

	type args struct {
		target any
	}

	tests := []struct {
		name        string
		args        args
		wantErr     bool
		errType     error
		validateOut func(t *testing.T, s *Schema)
	}{
		{
			name:    "nil interface",
			args:    args{target: nil},
			wantErr: true,
			errType: errs.InvalidTarget,
		},
		{
			name:    "non-pointer target",
			args:    args{target: struct{}{}},
			wantErr: true,
			errType: errs.InvalidTarget,
		},
		{
			name:    "nil pointer target",
			args:    args{target: (*struct{})(nil)},
			wantErr: true,
			errType: errs.InvalidTarget,
		},
		{
			name:    "pointer to non-struct",
			args:    args{target: new(int)},
			wantErr: true,
			errType: errs.InvalidTarget,
		},
		{
			name:    "collects exported and nested fields",
			args:    args{target: &config{}},
			wantErr: false,
			errType: nil,
			validateOut: func(t *testing.T, s *Schema) {
				t.Helper()

				if got, want := len(s.Fields), 4; got != want {
					t.Fatalf("Build() field count = %d, want %d", got, want)
				}

				f0 := s.Fields[0]
				if f0.Path != "Name" {
					t.Fatalf("field[0].Path = %q, want %q", f0.Path, "Name")
				}
				if f0.ConfName != "name" || f0.EnvName != "NAME" {
					t.Fatalf("field[0] tags not parsed correctly: conf=%q env=%q", f0.ConfName, f0.EnvName)
				}
				if !f0.HasDefaultValue || f0.DefaultValue != "app" {
					t.Fatalf("field[0] default parsing incorrect: has=%v value=%q", f0.HasDefaultValue, f0.DefaultValue)
				}
				if !f0.Required {
					t.Fatalf("field[0] required parsing incorrect")
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
				if f3.EnvName != "FLAG" || !f3.Required {
					t.Fatalf("field[3] tags not parsed correctly: env=%q required=%v", f3.EnvName, f3.Required)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := Build(tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Build() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if s != nil {
					t.Fatalf("Build() expected nil schema on error, got %#v", s)
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Fatalf("Build() error = %v, want wrapped %v", err, tt.errType)
				}
				return
			}

			if s == nil {
				t.Fatalf("Build() got nil schema")
			}
			if tt.validateOut != nil {
				tt.validateOut(t, s)
			}
		})
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
