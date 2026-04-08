package defaults

import (
	"errors"
	"reflect"
	"testing"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

func TestApply(t *testing.T) {
	type args struct {
		sc *schema.Schema
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		errType error
	}{
		{
			name: "apply defaults to nil schema",
			args: args{
				sc: nil,
			},
			wantErr: true,
			errType: errs.InvalidSchemaNil,
		},
		{
			name: "apply defaults to empty schema",
			args: args{
				sc: &schema.Schema{},
			},
			wantErr: false,
			errType: nil,
		},
		{
			name: "apply defaults to schema with no default values",
			args: args{
				sc: &schema.Schema{
					Fields: []schema.Field{
						{
							GoName: "field1",
							Type:   reflect.TypeOf(""),
							Value:  reflect.ValueOf(new(string)).Elem(),
						},
						{
							GoName: "field2",
							Type:   reflect.TypeOf(0),
							Value:  reflect.ValueOf(new(int)).Elem(),
						},
					},
				},
			},
			wantErr: false,
			errType: nil,
		},
		{
			name: "apply defaults to schema with valid default values",
			args: args{
				sc: &schema.Schema{
					Fields: []schema.Field{
						{
							GoName:       "field1",
							Type:         reflect.TypeOf(""),
							Value:        reflect.ValueOf(new(string)).Elem(),
							DefaultValue: "default1",
						},
						{
							GoName:       "field2",
							Type:         reflect.TypeOf(0),
							Value:        reflect.ValueOf(new(int)).Elem(),
							DefaultValue: "42",
						},
					},
				},
			},
			wantErr: false,
			errType: nil,
		},
		{
			name: "apply defaults to schema with invalid default value",
			args: args{
				sc: &schema.Schema{
					Fields: []schema.Field{
						{
							GoName:       "field1",
							Type:         reflect.TypeOf(""),
							Value:        reflect.ValueOf(new(string)).Elem(),
							DefaultValue: "default1",
						},
						{
							GoName:       "field2",
							Type:         reflect.TypeOf(0),
							Value:        reflect.ValueOf(new(int)).Elem(),
							DefaultValue: "not-an-int",
						},
					},
				},
			},
			wantErr: true,
			errType: errs.DecodeInvalidInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Apply(tt.args.sc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !errors.Is(err, tt.errType) {
				t.Errorf("Apply() error = %v, expected error type %v", err, tt.errType)
			}
		})
	}
}
