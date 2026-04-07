package decode

import (
	"encoding"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

var (
	durationType        = reflect.TypeOf(time.Duration(0))
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

func SetFieldValue(field schema.Field, raw any) error {
	if !field.Value.CanSet() {
		return errs.DecodeFieldCannotSet
	}

	return setValue(field.Value, raw)
}

func setValue(dst reflect.Value, raw any) error {
	if dst.Type() == durationType {
		return setDurationValue(dst, raw)
	}

	if canDecodeWithTextUnmarshaler(dst) {
		return setWithTextUnmarshaler(dst, raw)
	}

	switch dst.Kind() {
	case reflect.String:
		return setStringValue(dst, raw)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setIntValue(dst, raw)

	case reflect.Bool:
		return setBoolValue(dst, raw)

	case reflect.Float32, reflect.Float64:
		return setFloatValue(dst, raw)

	case reflect.Slice:
		return setSliceValue(dst, raw)

	default:
		return fmt.Errorf("%w: %s", errs.DecodeUnsupported, dst.Type())
	}
}

func setDurationValue(dst reflect.Value, raw any) error {
	v, err := toDuration(raw)
	if err != nil {
		return err
	}
	dst.SetInt(int64(v))
	return nil
}

func setStringValue(dst reflect.Value, raw any) error {
	v, ok := raw.(string)
	if !ok {
		return fmt.Errorf("%w: expected string, got %T", errs.DecodeTypeMismatch, raw)
	}
	dst.SetString(v)
	return nil
}

func setIntValue(dst reflect.Value, raw any) error {
	v, err := toInt64(raw)
	if err != nil {
		return err
	}
	if dst.OverflowInt(v) {
		return fmt.Errorf("%w: %v overflows %s", errs.DecodeInvalidInt, v, dst.Type())
	}
	dst.SetInt(v)
	return nil
}

func setBoolValue(dst reflect.Value, raw any) error {
	v, err := toBool(raw)
	if err != nil {
		return err
	}
	dst.SetBool(v)
	return nil
}

func setFloatValue(dst reflect.Value, raw any) error {
	v, err := toFloat64(raw)
	if err != nil {
		return err
	}
	if dst.OverflowFloat(v) {
		return fmt.Errorf("%w: %v overflows %s", errs.DecodeInvalidFloat, v, dst.Type())
	}
	dst.SetFloat(v)
	return nil
}

func setSliceValue(dst reflect.Value, raw any) error {
	v, err := toSlice(dst.Type(), raw)
	if err != nil {
		return err
	}
	dst.Set(v)
	return nil
}

func canDecodeWithTextUnmarshaler(v reflect.Value) bool {
	if v.CanAddr() && v.Addr().Type().Implements(textUnmarshalerType) {
		return true
	}
	return v.Type().Implements(textUnmarshalerType)
}

func setWithTextUnmarshaler(v reflect.Value, raw any) error {
	s, ok := raw.(string)
	if !ok {
		return fmt.Errorf("%w: expected string, got %T", errs.DecodeTypeMismatch, raw)
	}

	if v.CanAddr() && v.Addr().Type().Implements(textUnmarshalerType) {
		u := v.Addr().Interface().(encoding.TextUnmarshaler)
		if err := u.UnmarshalText([]byte(s)); err != nil {
			return errs.WrapDecode(errs.Decode, "", err)
		}
		return nil
	}

	u := v.Interface().(encoding.TextUnmarshaler)
	if err := u.UnmarshalText([]byte(s)); err != nil {
		return errs.WrapDecode(errs.Decode, "", err)
	}
	return nil
}

func toDuration(raw any) (time.Duration, error) {
	if v, ok := raw.(time.Duration); ok {
		return v, nil
	}
	if s, ok := raw.(string); ok {
		d, err := time.ParseDuration(s)
		if err != nil {
			return 0, fmt.Errorf("%w: %q", errs.DecodeInvalidDuration, s)
		}
		return d, nil
	}
	if n, ok, err := toInt64FromNumeric(raw, errs.DecodeInvalidDuration, "duration"); ok || err != nil {
		return time.Duration(n), err
	}
	return 0, fmt.Errorf("%w: expected duration, got %T", errs.DecodeTypeMismatch, raw)
}

func toInt64(raw any) (int64, error) {
	if s, ok := raw.(string); ok {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: %q", errs.DecodeInvalidInt, s)
		}
		return i, nil
	}
	if n, ok, err := toInt64FromNumeric(raw, errs.DecodeInvalidInt, "int"); ok || err != nil {
		return n, err
	}
	return 0, fmt.Errorf("%w: expected int, got %T", errs.DecodeTypeMismatch, raw)
}

func toInt64FromNumeric(raw any, invalidErr error, target string) (int64, bool, error) {
	switch v := raw.(type) {
	case int:
		return int64(v), true, nil
	case int8:
		return int64(v), true, nil
	case int16:
		return int64(v), true, nil
	case int32:
		return int64(v), true, nil
	case int64:
		return v, true, nil
	case uint:
		if uint64(v) > uint64(math.MaxInt64) {
			return 0, true, fmt.Errorf("%w: %v overflows int64", invalidErr, v)
		}
		return int64(v), true, nil
	case uint8:
		return int64(v), true, nil
	case uint16:
		return int64(v), true, nil
	case uint32:
		return int64(v), true, nil
	case uint64:
		if v > uint64(math.MaxInt64) {
			return 0, true, fmt.Errorf("%w: %v overflows int64", invalidErr, v)
		}
		return int64(v), true, nil
	case float32:
		i, err := floatToInt64(float64(v), invalidErr, target)
		return i, true, err
	case float64:
		i, err := floatToInt64(v, invalidErr, target)
		return i, true, err
	default:
		return 0, false, nil
	}
}

func floatToInt64(v float64, invalidErr error, target string) (int64, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("%w: %v", invalidErr, v)
	}
	if math.Trunc(v) != v {
		return 0, fmt.Errorf("%w: cannot convert non-integer float %v to %s", invalidErr, v, target)
	}
	if v < math.MinInt64 || v > math.MaxInt64 {
		return 0, fmt.Errorf("%w: %v overflows int64", invalidErr, v)
	}
	return int64(v), nil
}

func toBool(raw any) (bool, error) {
	switch v := raw.(type) {
	case bool:
		return v, nil
	case string:
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return false, fmt.Errorf("%w: %q", errs.DecodeInvalidBool, v)
		}
		return parsed, nil
	default:
		return false, fmt.Errorf("%w: expected bool, got %T", errs.DecodeTypeMismatch, raw)
	}
}

func toFloat64(raw any) (float64, error) {
	switch v := raw.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: %q", errs.DecodeInvalidFloat, v)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("%w: expected float, got %T", errs.DecodeTypeMismatch, raw)
	}
}

func toSlice(targetType reflect.Type, raw any) (reflect.Value, error) {
	if raw == nil {
		return reflect.Zero(targetType), nil
	}

	rv := reflect.ValueOf(raw)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return reflect.Value{}, fmt.Errorf("%w: expected %s, got %T", errs.DecodeTypeMismatch, targetType, raw)
	}

	out := reflect.MakeSlice(targetType, rv.Len(), rv.Len())
	for i := 0; i < rv.Len(); i++ {
		if err := setValue(out.Index(i), rv.Index(i).Interface()); err != nil {
			return reflect.Value{}, errs.WrapDecode(errs.Decode, fmt.Sprintf("index %d", i), err)
		}
	}

	return out, nil
}
