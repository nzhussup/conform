package konform

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/nzhussup/konform/internal/decode"
	internalschema "github.com/nzhussup/konform/internal/schema"
)

func TestLoad(t *testing.T) {
	type cfg struct {
		Name string `env:"NAME" required:"true"`
		Port int    `default:"8080" env:"PORT"`
	}

	optionWithSource := func(loader sourceLoader) Option {
		return func(o *loadOptions) error {
			if o == nil {
				return ErrInvalidSchema
			}
			o.sources = append(o.sources, loader)
			return nil
		}
	}

	tests := []struct {
		name        string
		target      any
		opts        func(t *testing.T) []Option
		wantErrType error
		wantErrLike []string
		validate    func(t *testing.T, target *cfg, err error)
	}{
		{
			name:        "invalid target",
			target:      cfg{},
			opts:        func(t *testing.T) []Option { t.Helper(); return nil },
			wantErrType: ErrInvalidTarget,
		},
		{
			name:   "option error is returned",
			target: &cfg{},
			opts: func(t *testing.T) []Option {
				t.Helper()
				return []Option{
					func(_ *loadOptions) error { return fmt.Errorf("option failed") },
				}
			},
			wantErrLike: []string{"option failed"},
		},
		{
			name:   "source error is returned",
			target: &cfg{},
			opts: func(t *testing.T) []Option {
				t.Helper()
				return []Option{
					optionWithSource(func(_ *internalschema.Schema) error {
						return fmt.Errorf("source failed")
					}),
				}
			},
			wantErrLike: []string{"source failed"},
		},
		{
			name:   "validation error for missing required",
			target: &cfg{},
			opts:   func(t *testing.T) []Option { t.Helper(); return nil },
			validate: func(t *testing.T, _ *cfg, err error) {
				t.Helper()
				var vErr *ValidationError
				if !errors.As(err, &vErr) {
					t.Fatalf("Load() error = %v, want ValidationError", err)
				}
				if !strings.Contains(err.Error(), "Name: required") {
					t.Fatalf("Load() error = %q, want to contain %q", err.Error(), "Name: required")
				}
			},
			wantErrType: ErrValidation,
		},
		{
			name:   "success with default then source override",
			target: &cfg{},
			opts: func(t *testing.T) []Option {
				t.Helper()
				return []Option{
					optionWithSource(func(sc *internalschema.Schema) error {
						for _, f := range sc.Fields {
							switch f.Path {
							case "Name":
								if err := decode.SetFieldValue(f, "svc"); err != nil {
									return err
								}
							case "Port":
								if err := decode.SetFieldValue(f, "9090"); err != nil {
									return err
								}
							}
						}
						return nil
					}),
				}
			},
			validate: func(t *testing.T, target *cfg, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("Load() error = %v, want nil", err)
				}
				if target.Name != "svc" {
					t.Fatalf("Name = %q, want %q", target.Name, "svc")
				}
				if target.Port != 9090 {
					t.Fatalf("Port = %d, want %d", target.Port, 9090)
				}
			},
		},
		{
			name:   "success with env source",
			target: &cfg{},
			opts: func(t *testing.T) []Option {
				t.Helper()
				t.Setenv("NAME", "api")
				t.Setenv("PORT", "7777")
				return []Option{FromEnv()}
			},
			validate: func(t *testing.T, target *cfg, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("Load() error = %v, want nil", err)
				}
				if target.Name != "api" {
					t.Fatalf("Name = %q, want %q", target.Name, "api")
				}
				if target.Port != 7777 {
					t.Fatalf("Port = %d, want %d", target.Port, 7777)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.opts(t)
			err := Load(tt.target, opts...)
			if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
				t.Fatalf("Load() error = %v, want wrapped %v", err, tt.wantErrType)
			}
			for _, part := range tt.wantErrLike {
				if err == nil || !strings.Contains(err.Error(), part) {
					t.Fatalf("Load() error = %v, want to contain %q", err, part)
				}
			}
			if tt.validate != nil {
				cfgTarget, _ := tt.target.(*cfg)
				tt.validate(t, cfgTarget, err)
				return
			}
			if tt.wantErrType == nil && len(tt.wantErrLike) == 0 && err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}
		})
	}
}
