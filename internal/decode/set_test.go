package decode

import (
	"errors"
	"reflect"
	"testing"

	"github.com/nzhussup/conform/internal/errs"
	"github.com/nzhussup/conform/internal/schema"
)

func TestSet(t *testing.T) {
	type args struct {
		field *schema.Field
		raw   string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		errType error
	}{
		{
			name: "set string field",
			args: args{
				field: &schema.Field{
					Type:  reflect.TypeOf(""),
					Value: reflect.ValueOf(new(string)).Elem(),
				},
				raw: "hello",
			},
			wantErr: false,
			errType: nil,
		},
		{
			name: "set int field",
			args: args{
				field: &schema.Field{
					Type:  reflect.TypeOf(0),
					Value: reflect.ValueOf(new(int)).Elem(),
				},
				raw: "42",
			},
			wantErr: false,
			errType: nil,
		},
		{
			name: "set bool field",
			args: args{
				field: &schema.Field{
					Type:  reflect.TypeOf(true),
					Value: reflect.ValueOf(new(bool)).Elem(),
				},
				raw: "true",
			},
			wantErr: false,
			errType: nil,
		},
		{
			name: "invalid int value",
			args: args{
				field: &schema.Field{
					Type:  reflect.TypeOf(0),
					Value: reflect.ValueOf(new(int)).Elem(),
				},
				raw: "not-an-int",
			},
			wantErr: true,
			errType: errs.DecodeInvalidInt,
		},
		{
			name: "invalid bool value",
			args: args{
				field: &schema.Field{
					Type:  reflect.TypeOf(true),
					Value: reflect.ValueOf(new(bool)).Elem(),
				},
				raw: "not-a-bool",
			},
			wantErr: true,
			errType: errs.DecodeInvalidBool,
		},
		{
			name: "unsupported field type",
			args: args{
				field: &schema.Field{
					Type:  reflect.TypeOf(3.14),
					Value: reflect.ValueOf(new(float64)).Elem(),
				},
				raw: "3.14",
			},
			wantErr: true,
			errType: errs.DecodeUnsupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetFieldValue(*tt.args.field, tt.args.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SetFieldValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("SetFieldValue() error = %v, want wrapped %v", err, tt.errType)
			}
		})
	}
}
