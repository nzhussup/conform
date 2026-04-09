package validators

import (
	"reflect"
	"strings"
)

func numericAsFloat64(v reflect.Value) (float64, bool) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(v.Uint()), true
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	default:
		return 0, false
	}
}

func parsePipeSeparated(s string) []string {
	var result []string
	for _, part := range strings.Split(s, "|") {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}

func isNumericKind(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Float64
}
