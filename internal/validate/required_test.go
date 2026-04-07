package validate

import (
	"errors"
	"reflect"
	"testing"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

func TestMissingRequired(t *testing.T) {
	tests := []struct {
		name        string
		sc          *schema.Schema
		wantMissing []string
		wantErrType error
	}{
		{
			name:        "nil schema",
			sc:          nil,
			wantErrType: errs.InvalidSchemaNil,
		},
		{
			name: "no required fields",
			sc: &schema.Schema{
				Fields: []schema.Field{
					{Path: "Port", Required: false, Value: reflect.ValueOf(new(int)).Elem()},
				},
			},
			wantMissing: []string{},
		},
		{
			name: "required zero value is missing",
			sc: func() *schema.Schema {
				var name string
				return &schema.Schema{
					Fields: []schema.Field{
						{Path: "Name", Required: true, Value: reflect.ValueOf(&name).Elem()},
					},
				}
			}(),
			wantMissing: []string{"Name"},
		},
		{
			name: "required non-zero is not missing",
			sc: func() *schema.Schema {
				name := "svc"
				return &schema.Schema{
					Fields: []schema.Field{
						{Path: "Name", Required: true, Value: reflect.ValueOf(&name).Elem()},
					},
				}
			}(),
			wantMissing: []string{},
		},
		{
			name: "mixed fields return only missing required",
			sc: func() *schema.Schema {
				var name string
				port := 8080
				return &schema.Schema{
					Fields: []schema.Field{
						{Path: "Name", Required: true, Value: reflect.ValueOf(&name).Elem()},
						{Path: "Port", Required: true, Value: reflect.ValueOf(&port).Elem()},
						{Path: "Optional", Required: false, Value: reflect.ValueOf(new(string)).Elem()},
					},
				}
			}(),
			wantMissing: []string{"Name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MissingRequired(tt.sc)
			if tt.wantErrType != nil {
				if err == nil {
					t.Fatalf("MissingRequired() error = nil, want %v", tt.wantErrType)
				}
				if !errors.Is(err, tt.wantErrType) {
					t.Fatalf("MissingRequired() error = %v, want wrapped %v", err, tt.wantErrType)
				}
				return
			}
			if err != nil {
				t.Fatalf("MissingRequired() error = %v, want nil", err)
			}

			paths := make([]string, 0, len(got))
			for _, f := range got {
				paths = append(paths, f.Path)
			}
			if !reflect.DeepEqual(paths, tt.wantMissing) {
				t.Fatalf("MissingRequired() paths = %#v, want %#v", paths, tt.wantMissing)
			}
		})
	}
}
