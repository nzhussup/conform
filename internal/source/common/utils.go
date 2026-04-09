package common

import (
	"path/filepath"
	"reflect"
	"strings"

	"github.com/nzhussup/konform/internal/decode"
	"github.com/nzhussup/konform/internal/schema"
)

func resolvePath(path string, callerDir string) string {
	if filepath.IsAbs(path) || callerDir == "" {
		return path
	}
	return filepath.Join(callerDir, path)
}

func defaultSourceLabel(path string, format string) string {
	if path != "" {
		return path
	}
	return format
}

func isStructField(field schema.Field) bool {
	return field.Type.Kind() == reflect.Struct
}

func setFieldFromValue(field schema.Field, value any) error {
	return decode.SetFieldValue(field, value)
}

func BuildPathAliases(sc *schema.Schema) map[string]string {
	aliases := make(map[string]string, len(sc.Fields))
	for _, field := range sc.Fields {
		if field.KeyName == "" {
			continue
		}
		aliases[field.Path] = field.KeyName
	}
	return aliases
}

func ResolveLookupPath(field schema.Field, pathAliases map[string]string) string {
	if field.KeyName != "" {
		return field.KeyName
	}

	resolved := field.Path
	parts := strings.Split(field.Path, ".")

	for i := range parts {
		prefix := joinPath(parts[:i+1])
		alias, ok := pathAliases[prefix]
		if !ok {
			continue
		}

		suffix := joinPath(parts[i+1:])
		if suffix == "" {
			resolved = alias
		} else {
			resolved = alias + "." + suffix
		}
	}

	return resolved
}

func joinPath(parts []string) string {
	return strings.Join(parts, ".")
}

func GetByPath(doc Document, path string) (any, bool) {
	if path == "" {
		return nil, false
	}

	keys := strings.Split(path, ".")
	current := any(doc)

	for _, key := range keys {
		m, ok := asStringMap(current)
		if !ok {
			return nil, false
		}
		current, ok = m[key]
		if !ok {
			return nil, false
		}
	}

	return current, true
}

func asStringMap(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case Document:
		return map[string]any(m), true
	case map[string]any:
		return m, true
	case map[any]any:
		out := make(map[string]any, len(m))
		for k, val := range m {
			key, ok := k.(string)
			if !ok {
				return nil, false
			}
			out[key] = val
		}
		return out, true
	default:
		return nil, false
	}
}
