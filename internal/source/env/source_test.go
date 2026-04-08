package env

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

func TestLoad(t *testing.T) {
	const (
		envPort    = "KONFORM_TEST_ENV_PORT"
		envDebug   = "KONFORM_TEST_ENV_DEBUG"
		envTimeout = "KONFORM_TEST_ENV_TIMEOUT"
	)

	type setupOut struct {
		sc       *schema.Schema
		portPtr  *int
		debugPtr *bool
		durPtr   *time.Duration
	}

	tests := []struct {
		name        string
		setup       func(t *testing.T) setupOut
		wantErrType error
		wantErrLike []string
		validate    func(t *testing.T, out setupOut)
	}{
		{
			name: "nil schema",
			setup: func(t *testing.T) setupOut {
				t.Helper()
				return setupOut{sc: nil}
			},
			wantErrType: errs.InvalidSchemaNil,
		},
		{
			name: "missing env vars are skipped",
			setup: func(t *testing.T) setupOut {
				t.Helper()
				port := 0
				sc := &schema.Schema{
					Fields: []schema.Field{
						{
							Path:    "Port",
							EnvName: envPort,
							Type:    reflect.TypeOf(0),
							Value:   reflect.ValueOf(&port).Elem(),
						},
					},
				}
				return setupOut{sc: sc, portPtr: &port}
			},
			validate: func(t *testing.T, out setupOut) {
				t.Helper()
				if *out.portPtr != 0 {
					t.Fatalf("Port = %d, want 0", *out.portPtr)
				}
			},
		},
		{
			name: "decodes supported scalar types from env strings",
			setup: func(t *testing.T) setupOut {
				t.Helper()
				t.Setenv(envPort, "9090")
				t.Setenv(envDebug, "true")
				t.Setenv(envTimeout, "1500ms")

				port := 0
				debug := false
				timeout := time.Duration(0)
				sc := &schema.Schema{
					Fields: []schema.Field{
						{
							Path:    "Port",
							EnvName: envPort,
							Type:    reflect.TypeOf(0),
							Value:   reflect.ValueOf(&port).Elem(),
						},
						{
							Path:    "Debug",
							EnvName: envDebug,
							Type:    reflect.TypeOf(true),
							Value:   reflect.ValueOf(&debug).Elem(),
						},
						{
							Path:    "Timeout",
							EnvName: envTimeout,
							Type:    reflect.TypeOf(time.Duration(0)),
							Value:   reflect.ValueOf(&timeout).Elem(),
						},
					},
				}
				return setupOut{sc: sc, portPtr: &port, debugPtr: &debug, durPtr: &timeout}
			},
			validate: func(t *testing.T, out setupOut) {
				t.Helper()
				if *out.portPtr != 9090 {
					t.Fatalf("Port = %d, want 9090", *out.portPtr)
				}
				if *out.debugPtr != true {
					t.Fatalf("Debug = %v, want true", *out.debugPtr)
				}
				if *out.durPtr != 1500*time.Millisecond {
					t.Fatalf("Timeout = %v, want %v", *out.durPtr, 1500*time.Millisecond)
				}
			},
		},
		{
			name: "decode error includes env context",
			setup: func(t *testing.T) setupOut {
				t.Helper()
				t.Setenv(envDebug, "not-bool")

				debug := false
				sc := &schema.Schema{
					Fields: []schema.Field{
						{
							Path:    "Debug",
							EnvName: envDebug,
							Type:    reflect.TypeOf(true),
							Value:   reflect.ValueOf(&debug).Elem(),
						},
					},
				}
				return setupOut{sc: sc, debugPtr: &debug}
			},
			wantErrType: errs.Decode,
			wantErrLike: []string{`env "` + envDebug + `" -> Debug`, "invalid bool value"},
		},
		{
			name: "collects multiple env decode errors",
			setup: func(t *testing.T) setupOut {
				t.Helper()
				t.Setenv(envPort, "not-int")
				t.Setenv(envDebug, "not-bool")

				port := 0
				debug := false
				sc := &schema.Schema{
					Fields: []schema.Field{
						{
							Path:    "Port",
							EnvName: envPort,
							Type:    reflect.TypeOf(0),
							Value:   reflect.ValueOf(&port).Elem(),
						},
						{
							Path:    "Debug",
							EnvName: envDebug,
							Type:    reflect.TypeOf(true),
							Value:   reflect.ValueOf(&debug).Elem(),
						},
					},
				}
				return setupOut{sc: sc, portPtr: &port, debugPtr: &debug}
			},
			wantErrType: errs.Decode,
			wantErrLike: []string{
				`env "` + envPort + `" -> Port`,
				"invalid int value",
				`env "` + envDebug + `" -> Debug`,
				"invalid bool value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := tt.setup(t)
			err := Load(out.sc)
			if tt.wantErrType != nil {
				if err == nil {
					t.Fatalf("Load() error = nil, want %v", tt.wantErrType)
				}
				if !errors.Is(err, tt.wantErrType) {
					t.Fatalf("Load() error = %v, want wrapped %v", err, tt.wantErrType)
				}
				for _, part := range tt.wantErrLike {
					if !strings.Contains(err.Error(), part) {
						t.Fatalf("Load() error = %q, want to contain %q", err.Error(), part)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}
			if tt.validate != nil {
				tt.validate(t, out)
			}
		})
	}
}
