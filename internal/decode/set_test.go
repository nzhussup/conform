package decode

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

type upperText string

func (u *upperText) UnmarshalText(text []byte) error {
	*u = upperText(strings.ToUpper(string(text)))
	return nil
}

type failingText string

func (f *failingText) UnmarshalText(_ []byte) error {
	return fmt.Errorf("bad text")
}

func TestSetFieldValue(t *testing.T) {
	makeField := func(dst any) schema.Field {
		v := reflect.ValueOf(dst)
		return schema.Field{
			Type:  v.Elem().Type(),
			Value: v.Elem(),
		}
	}

	tests := []struct {
		name        string
		field       schema.Field
		raw         any
		want        any
		wantErrType error
		wantErrLike string
	}{
		{
			name:  "string from string",
			field: makeField(new(string)),
			raw:   "hello",
			want:  "hello",
		},
		{
			name:        "string rejects bool",
			field:       makeField(new(string)),
			raw:         true,
			wantErrType: errs.DecodeTypeMismatch,
			wantErrLike: "expected string, got bool",
		},
		{
			name:        "string rejects map",
			field:       makeField(new(string)),
			raw:         map[string]any{"a": "b"},
			wantErrType: errs.DecodeTypeMismatch,
			wantErrLike: "expected string, got map[string]interface {}",
		},
		{
			name:  "int from string",
			field: makeField(new(int)),
			raw:   "42",
			want:  int(42),
		},
		{
			name:  "int from numeric float integer",
			field: makeField(new(int)),
			raw:   42.0,
			want:  int(42),
		},
		{
			name:        "int rejects non-integer float",
			field:       makeField(new(int)),
			raw:         1.2,
			wantErrType: errs.DecodeInvalidInt,
			wantErrLike: "cannot convert non-integer float 1.2 to int",
		},
		{
			name:        "int rejects invalid string",
			field:       makeField(new(int)),
			raw:         "abc",
			wantErrType: errs.DecodeInvalidInt,
			wantErrLike: "\"abc\"",
		},
		{
			name:        "int rejects type mismatch",
			field:       makeField(new(int)),
			raw:         true,
			wantErrType: errs.DecodeTypeMismatch,
			wantErrLike: "expected int, got bool",
		},
		{
			name:        "int rejects uint64 overflow",
			field:       makeField(new(int64)),
			raw:         uint64(math.MaxInt64) + 1,
			wantErrType: errs.DecodeInvalidInt,
			wantErrLike: "overflows int64",
		},
		{
			name:        "int8 rejects overflow",
			field:       makeField(new(int8)),
			raw:         int64(128),
			wantErrType: errs.DecodeInvalidInt,
			wantErrLike: "overflows int8",
		},
		{
			name:  "bool from string",
			field: makeField(new(bool)),
			raw:   "true",
			want:  true,
		},
		{
			name:        "bool rejects invalid string",
			field:       makeField(new(bool)),
			raw:         "truthy",
			wantErrType: errs.DecodeInvalidBool,
			wantErrLike: "\"truthy\"",
		},
		{
			name:        "bool rejects type mismatch",
			field:       makeField(new(bool)),
			raw:         1,
			wantErrType: errs.DecodeTypeMismatch,
			wantErrLike: "expected bool, got int",
		},
		{
			name:  "float64 from string",
			field: makeField(new(float64)),
			raw:   "3.14",
			want:  float64(3.14),
		},
		{
			name:  "float64 from int",
			field: makeField(new(float64)),
			raw:   int(7),
			want:  float64(7),
		},
		{
			name:        "float32 rejects overflow",
			field:       makeField(new(float32)),
			raw:         1e40,
			wantErrType: errs.DecodeInvalidFloat,
			wantErrLike: "overflows float32",
		},
		{
			name:        "float rejects invalid string",
			field:       makeField(new(float64)),
			raw:         "not-float",
			wantErrType: errs.DecodeInvalidFloat,
			wantErrLike: "\"not-float\"",
		},
		{
			name:  "duration from string",
			field: makeField(new(time.Duration)),
			raw:   "1500ms",
			want:  1500 * time.Millisecond,
		},
		{
			name:  "duration from numeric",
			field: makeField(new(time.Duration)),
			raw:   int64(5),
			want:  5 * time.Nanosecond,
		},
		{
			name:        "duration rejects non-integer float",
			field:       makeField(new(time.Duration)),
			raw:         1.25,
			wantErrType: errs.DecodeInvalidDuration,
			wantErrLike: "cannot convert non-integer float 1.25 to duration",
		},
		{
			name:        "duration rejects invalid string",
			field:       makeField(new(time.Duration)),
			raw:         "later",
			wantErrType: errs.DecodeInvalidDuration,
			wantErrLike: "\"later\"",
		},
		{
			name:        "duration rejects overflow",
			field:       makeField(new(time.Duration)),
			raw:         uint64(math.MaxInt64) + 1,
			wantErrType: errs.DecodeInvalidDuration,
			wantErrLike: "overflows int64",
		},
		{
			name:  "text unmarshaler from string",
			field: makeField(new(upperText)),
			raw:   "hello",
			want:  upperText("HELLO"),
		},
		{
			name:  "pointer text unmarshaler from string",
			field: makeField(new(*upperText)),
			raw:   "hello",
			want: func() any {
				v := upperText("HELLO")
				return &v
			}(),
		},
		{
			name:        "text unmarshaler rejects non-string",
			field:       makeField(new(upperText)),
			raw:         1,
			wantErrType: errs.DecodeTypeMismatch,
			wantErrLike: "expected string, got int",
		},
		{
			name:        "text unmarshaler returns decode error",
			field:       makeField(new(failingText)),
			raw:         "x",
			wantErrType: errs.Decode,
			wantErrLike: "bad text",
		},
		{
			name:        "pointer text unmarshaler returns decode error",
			field:       makeField(new(*failingText)),
			raw:         "x",
			wantErrType: errs.Decode,
			wantErrLike: "bad text",
		},
		{
			name:  "slice string from []any",
			field: makeField(new([]string)),
			raw:   []any{"a", "b"},
			want:  []string{"a", "b"},
		},
		{
			name:  "slice int with safe coercion",
			field: makeField(new([]int)),
			raw:   []any{"1", 2, 3.0},
			want:  []int{1, 2, 3},
		},
		{
			name:  "slice duration from mixed values",
			field: makeField(new([]time.Duration)),
			raw:   []any{"100ms", int64(2)},
			want:  []time.Duration{100 * time.Millisecond, 2 * time.Nanosecond},
		},
		{
			name:  "slice text unmarshaler",
			field: makeField(new([]upperText)),
			raw:   []any{"json", "TEXT"},
			want:  []upperText{"JSON", "TEXT"},
		},
		{
			name:        "slice rejects scalar",
			field:       makeField(new([]int)),
			raw:         "1,2,3",
			wantErrType: errs.DecodeTypeMismatch,
			wantErrLike: "expected []int, got string",
		},
		{
			name:        "slice element conversion error includes index",
			field:       makeField(new([]int)),
			raw:         []any{1, 2.2},
			wantErrType: errs.DecodeInvalidInt,
			wantErrLike: "index 1",
		},
		{
			name:        "unsupported kind",
			field:       makeField(new(struct{})),
			raw:         "{}",
			wantErrType: errs.DecodeUnsupported,
		},
		{
			name: "field cannot set",
			field: schema.Field{
				Type:  reflect.TypeOf(""),
				Value: reflect.ValueOf(""),
			},
			raw:         "x",
			wantErrType: errs.DecodeFieldCannotSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetFieldValue(tt.field, tt.raw)
			if tt.wantErrType == nil {
				if err != nil {
					t.Fatalf("SetFieldValue() error = %v, want nil", err)
				}
				got := tt.field.Value.Interface()
				if !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("SetFieldValue() value = %#v, want %#v", got, tt.want)
				}
				return
			}

			if err == nil {
				t.Fatalf("SetFieldValue() error = nil, want %v", tt.wantErrType)
			}
			if !errors.Is(err, tt.wantErrType) {
				t.Fatalf("SetFieldValue() error = %v, want wrapped %v", err, tt.wantErrType)
			}
			if tt.wantErrLike != "" && !strings.Contains(err.Error(), tt.wantErrLike) {
				t.Fatalf("SetFieldValue() error = %q, want to contain %q", err.Error(), tt.wantErrLike)
			}
		})
	}
}

func TestDecodeHelpers(t *testing.T) {
	t.Run("toDuration handles direct duration and type mismatch", func(t *testing.T) {
		got, err := toDuration(2 * time.Second)
		if err != nil {
			t.Fatalf("toDuration(duration) error = %v, want nil", err)
		}
		if got != 2*time.Second {
			t.Fatalf("toDuration(duration) = %v, want %v", got, 2*time.Second)
		}

		_, err = toDuration(true)
		if err == nil {
			t.Fatalf("toDuration(bool) error = nil, want error")
		}
		if !errors.Is(err, errs.DecodeTypeMismatch) {
			t.Fatalf("toDuration(bool) error = %v, want wrapped %v", err, errs.DecodeTypeMismatch)
		}
	})

	t.Run("toInt64FromNumeric handles additional numeric kinds", func(t *testing.T) {
		cases := []any{int(1), int8(2), int16(3), int32(4), int64(5), uint8(7), uint16(8), uint32(9), float32(10)}
		for _, in := range cases {
			got, ok, err := toInt64FromNumeric(in, errs.DecodeInvalidInt, "int")
			if err != nil {
				t.Fatalf("toInt64FromNumeric(%T) error = %v, want nil", in, err)
			}
			if !ok {
				t.Fatalf("toInt64FromNumeric(%T) ok = false, want true", in)
			}
			if got <= 0 {
				t.Fatalf("toInt64FromNumeric(%T) = %d, want > 0", in, got)
			}
		}
	})

	t.Run("toInt64FromNumeric handles uint and uint64", func(t *testing.T) {
		for _, in := range []any{uint(11), uint64(12)} {
			got, ok, err := toInt64FromNumeric(in, errs.DecodeInvalidInt, "int")
			if err != nil {
				t.Fatalf("toInt64FromNumeric(%T) error = %v, want nil", in, err)
			}
			if !ok || got <= 0 {
				t.Fatalf("toInt64FromNumeric(%T) = (%d, %v), want (>0, true)", in, got, ok)
			}
		}
	})

	t.Run("toInt64FromNumeric overflow paths", func(t *testing.T) {
		_, ok, err := toInt64FromNumeric(uint64(math.MaxInt64)+1, errs.DecodeInvalidInt, "int")
		if !ok || err == nil || !errors.Is(err, errs.DecodeInvalidInt) {
			t.Fatalf("uint64 overflow = (ok=%v, err=%v), want wrapped %v", ok, err, errs.DecodeInvalidInt)
		}

		_, ok, err = toInt64FromNumeric(float64(math.MaxInt64)*2, errs.DecodeInvalidInt, "int")
		if !ok || err == nil || !errors.Is(err, errs.DecodeInvalidInt) {
			t.Fatalf("float overflow = (ok=%v, err=%v), want wrapped %v", ok, err, errs.DecodeInvalidInt)
		}
	})

	t.Run("toInt64FromNumeric returns not-handled for unsupported kind", func(t *testing.T) {
		got, ok, err := toInt64FromNumeric("11", errs.DecodeInvalidInt, "int")
		if err != nil {
			t.Fatalf("toInt64FromNumeric(string) error = %v, want nil", err)
		}
		if ok {
			t.Fatalf("toInt64FromNumeric(string) ok = true, want false")
		}
		if got != 0 {
			t.Fatalf("toInt64FromNumeric(string) = %d, want 0", got)
		}
	})

	t.Run("floatToInt64 rejects NaN and Inf", func(t *testing.T) {
		for _, in := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
			_, err := floatToInt64(in, errs.DecodeInvalidInt, "int")
			if err == nil {
				t.Fatalf("floatToInt64(%v) error = nil, want error", in)
			}
			if !errors.Is(err, errs.DecodeInvalidInt) {
				t.Fatalf("floatToInt64(%v) error = %v, want wrapped %v", in, err, errs.DecodeInvalidInt)
			}
		}
	})

	t.Run("toBool accepts bool directly", func(t *testing.T) {
		got, err := toBool(true)
		if err != nil {
			t.Fatalf("toBool(true) error = %v, want nil", err)
		}
		if !got {
			t.Fatalf("toBool(true) = false, want true")
		}
	})

	t.Run("toFloat64 handles all numeric kinds and bool mismatch", func(t *testing.T) {
		cases := []struct {
			in   any
			want float64
		}{
			{float32(1.5), 1.5},
			{float64(2.5), 2.5},
			{int(3), 3},
			{int8(4), 4},
			{int16(5), 5},
			{int32(6), 6},
			{int64(7), 7},
			{uint(8), 8},
			{uint8(9), 9},
			{uint16(10), 10},
			{uint32(11), 11},
			{uint64(12), 12},
		}

		for _, c := range cases {
			got, err := toFloat64(c.in)
			if err != nil {
				t.Fatalf("toFloat64(%T) error = %v, want nil", c.in, err)
			}
			if got != c.want {
				t.Fatalf("toFloat64(%T) = %v, want %v", c.in, got, c.want)
			}
		}

		_, err := toFloat64(true)
		if err == nil {
			t.Fatalf("toFloat64(bool) error = nil, want error")
		}
		if !errors.Is(err, errs.DecodeTypeMismatch) {
			t.Fatalf("toFloat64(bool) error = %v, want wrapped %v", err, errs.DecodeTypeMismatch)
		}
	})

	t.Run("toSlice returns zero slice for nil input", func(t *testing.T) {
		v, err := toSlice(reflect.TypeOf([]int{}), nil)
		if err != nil {
			t.Fatalf("toSlice(nil) error = %v, want nil", err)
		}
		if !v.IsNil() {
			t.Fatalf("toSlice(nil) = %#v, want nil slice", v.Interface())
		}
	})

	t.Run("canDecodeWithTextUnmarshaler false for plain type", func(t *testing.T) {
		if canDecodeWithTextUnmarshaler(reflect.ValueOf(1)) {
			t.Fatalf("canDecodeWithTextUnmarshaler(int) = true, want false")
		}
	})
}
