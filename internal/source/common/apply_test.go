package common

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

func TestApply(t *testing.T) {
	type nested struct{}

	tests := []struct {
		name        string
		scBuilder   func() *schema.Schema
		doc         Document
		mode        UnknownKeySuggestionMode
		wantErrType error
		wantErrLike []string
		validate    func(t *testing.T, sc *schema.Schema)
	}{
		{
			name:        "nil schema",
			scBuilder:   func() *schema.Schema { return nil },
			doc:         Document{},
			mode:        UnknownKeySuggestionOff,
			wantErrType: errs.InvalidSchemaNil,
		},
		{
			name: "missing path is ignored",
			scBuilder: func() *schema.Schema {
				var port int
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:    "Port",
							KeyName: "port",
							Type:    reflect.TypeOf(0),
							Value:   reflect.ValueOf(&port).Elem(),
						},
					},
				}
			},
			mode: UnknownKeySuggestionOff,
			doc:  Document{},
			validate: func(t *testing.T, sc *schema.Schema) {
				t.Helper()
				if got := sc.Fields[0].Value.Interface().(int); got != 0 {
					t.Fatalf("port = %d, want 0", got)
				}
			},
		},
		{
			name: "decode error is wrapped with field context",
			scBuilder: func() *schema.Schema {
				var enabled string
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:    "Enabled",
							KeyName: "enabled",
							Type:    reflect.TypeOf(""),
							Value:   reflect.ValueOf(&enabled).Elem(),
						},
					},
				}
			},
			mode:        UnknownKeySuggestionOff,
			doc:         Document{"enabled": true},
			wantErrType: errs.DecodeSourceField,
			wantErrLike: []string{"yaml", "enabled", "Enabled", "expected string, got bool"},
		},
		{
			name: "collects multiple decode errors",
			scBuilder: func() *schema.Schema {
				var name string
				var port int
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:    "Name",
							KeyName: "name",
							Type:    reflect.TypeOf(""),
							Value:   reflect.ValueOf(&name).Elem(),
						},
						{
							Path:    "Port",
							KeyName: "port",
							Type:    reflect.TypeOf(0),
							Value:   reflect.ValueOf(&port).Elem(),
						},
					},
				}
			},
			mode:        UnknownKeySuggestionOff,
			doc:         Document{"name": true, "port": "not-int"},
			wantErrType: errs.DecodeSourceField,
			wantErrLike: []string{"yaml \"name\" -> Name", "expected string, got bool", "yaml \"port\" -> Port", "invalid int value"},
		},
		{
			name: "nested field uses parent alias",
			scBuilder: func() *schema.Schema {
				var parent nested
				var port int
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:    "Server",
							KeyName: "server_cfg",
							Type:    reflect.TypeOf(parent),
							Value:   reflect.ValueOf(&parent).Elem(),
						},
						{
							Path:  "Server.Port",
							Type:  reflect.TypeOf(0),
							Value: reflect.ValueOf(&port).Elem(),
						},
					},
				}
			},
			mode: UnknownKeySuggestionOff,
			doc: Document{
				"server_cfg": map[string]any{
					"Port": "9090",
				},
			},
			validate: func(t *testing.T, sc *schema.Schema) {
				t.Helper()
				if got := sc.Fields[1].Value.Interface().(int); got != 9090 {
					t.Fatalf("port = %d, want 9090", got)
				}
			},
		},
		{
			name: "explicit key typo reports unexpected config key with schema suggestion",
			scBuilder: func() *schema.Schema {
				var port int
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:    "AppName",
							KeyName: "App.Nam",
							Type:    reflect.TypeOf(0),
							Value:   reflect.ValueOf(&port).Elem(),
						},
					},
				}
			},
			mode: UnknownKeySuggestionError,
			doc: Document{
				"App": map[string]any{
					"Name": 8080,
				},
			},
			wantErrType: errs.DecodeSourceField,
			wantErrLike: []string{`unknown configuration key "App.Name"`, `did you mean "App.Nam"?`},
		},
		{
			name: "non explicit path typo reports unexpected config key with schema suggestion",
			scBuilder: func() *schema.Schema {
				type server struct{ Port int }
				var s server
				var port int
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:  "Server",
							Type:  reflect.TypeOf(s),
							Value: reflect.ValueOf(&s).Elem(),
						},
						{
							Path:  "Server.Port",
							Type:  reflect.TypeOf(0),
							Value: reflect.ValueOf(&port).Elem(),
						},
					},
				}
			},
			mode: UnknownKeySuggestionError,
			doc: Document{
				"Server": map[string]any{
					"Porrt": 8080,
				},
			},
			wantErrType: errs.DecodeSourceField,
			wantErrLike: []string{`unknown configuration key "Server.Porrt"`, `did you mean "Server.Port"?`},
		},
		{
			name: "unexpected dotted key suggests struct key",
			scBuilder: func() *schema.Schema {
				var appName string
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:  "AppName",
							Type:  reflect.TypeOf(""),
							Value: reflect.ValueOf(&appName).Elem(),
						},
					},
				}
			},
			mode: UnknownKeySuggestionError,
			doc: Document{
				"App": map[string]any{
					"Name": "konform",
				},
			},
			wantErrType: errs.DecodeSourceField,
			wantErrLike: []string{`unknown configuration key "App.Name"`, `did you mean "AppName"?`},
		},
		{
			name: "extra file keys are ignored in non-strict mode",
			scBuilder: func() *schema.Schema {
				var appDebug bool
				return &schema.Schema{
					Fields: []schema.Field{
						{
							Path:  "AppDebug",
							Type:  reflect.TypeOf(true),
							Value: reflect.ValueOf(&appDebug).Elem(),
						},
					},
				}
			},
			mode: UnknownKeySuggestionOff,
			doc: Document{
				"AppDebug": true,
				"App": map[string]any{
					"Debug": true,
				},
			},
			validate: func(t *testing.T, sc *schema.Schema) {
				t.Helper()
				if got := sc.Fields[0].Value.Interface().(bool); !got {
					t.Fatalf("AppDebug = %v, want true", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := tt.scBuilder()
			err := ApplyWithMode(sc, tt.doc, "yaml", tt.mode, "config.yaml")

			if tt.wantErrType == nil {
				if err != nil {
					t.Fatalf("Apply() error = %v, want nil", err)
				}
				if tt.validate != nil {
					tt.validate(t, sc)
				}
				return
			}

			if err == nil {
				t.Fatalf("Apply() error = nil, want %v", tt.wantErrType)
			}
			if !errors.Is(err, tt.wantErrType) {
				t.Fatalf("Apply() error = %v, want wrapped %v", err, tt.wantErrType)
			}
			for _, part := range tt.wantErrLike {
				if !strings.Contains(err.Error(), part) {
					t.Fatalf("Apply() error = %q, want to contain %q", err.Error(), part)
				}
			}
		})
	}
}

func TestApplyWrapperAndWarnMode(t *testing.T) {
	var appName string
	sc := &schema.Schema{
		Fields: []schema.Field{
			{
				Path:  "AppName",
				Type:  reflect.TypeOf(""),
				Value: reflect.ValueOf(&appName).Elem(),
			},
		},
	}

	doc := Document{
		"App": map[string]any{"Name": "konform"},
	}

	stderr := captureStderr(t, func() {
		if err := Apply(sc, doc, "yaml"); err != nil {
			t.Fatalf("Apply() error = %v, want nil", err)
		}
	})

	if !strings.Contains(stderr, `konform: warning: yaml: unknown configuration key "App.Name"`) {
		t.Fatalf("stderr = %q, want unknown key warning", stderr)
	}
	if appName != "" {
		t.Fatalf("AppName = %q, want empty because no matching key", appName)
	}
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stderr = w

	fn()

	_ = w.Close()
	os.Stderr = orig

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	_ = r.Close()
	return buf.String()
}
