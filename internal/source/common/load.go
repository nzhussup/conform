package common

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/nzhussup/conform/internal/decode"
	"github.com/nzhussup/conform/internal/errs"
	"github.com/nzhussup/conform/internal/schema"
)

type Document map[string]interface{}

type UnmarshalFunc func([]byte) (Document, error)

func LoadFile(sc *schema.Schema, path string, callerDir string, format string, unmarshal UnmarshalFunc) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	resolvedPath := resolvePath(path, callerDir)

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return fmt.Errorf("%w: %s file %q: %w", errs.DecodeSourceRead, format, path, err)
	}

	doc, err := unmarshal(data)
	if err != nil {
		return fmt.Errorf("%w: %s file: %w", errs.DecodeSourceParse, format, err)
	}

	return Apply(sc, doc, format)
}

func resolvePath(path string, callerDir string) string {
	if filepath.IsAbs(path) || callerDir == "" {
		return path
	}
	return filepath.Join(callerDir, path)
}

func Apply(sc *schema.Schema, doc Document, format string) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	pathAliases := BuildPathAliases(sc)

	for _, field := range sc.Fields {
		if isStructField(field) {
			continue
		}

		lookupPath := ResolveLookupPath(field, pathAliases)
		value, ok := GetByPath(doc, lookupPath)
		if !ok {
			continue
		}

		if err := setFieldFromValue(field, value); err != nil {
			return fmt.Errorf("%w: %s %q -> %s: %w", errs.DecodeSourceField, format, lookupPath, field.Path, err)
		}
	}

	return nil
}

func isStructField(field schema.Field) bool {
	return field.Type.Kind() == reflect.Struct
}

func setFieldFromValue(field schema.Field, value interface{}) error {
	return decode.SetFieldValue(field, value)
}

func BuildPathAliases(sc *schema.Schema) map[string]string {
	aliases := make(map[string]string, len(sc.Fields))
	for _, field := range sc.Fields {
		if field.ConfName == "" {
			continue
		}
		aliases[field.Path] = field.ConfName
	}
	return aliases
}

func ResolveLookupPath(field schema.Field, pathAliases map[string]string) string {
	if field.ConfName != "" {
		return field.ConfName
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

func GetByPath(doc Document, path string) (interface{}, bool) {
	if path == "" {
		return nil, false
	}

	keys := strings.Split(path, ".")
	current := interface{}(doc)

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

func asStringMap(v interface{}) (map[string]interface{}, bool) {
	switch m := v.(type) {
	case Document:
		return map[string]interface{}(m), true
	case map[string]interface{}:
		return m, true
	case map[interface{}]interface{}:
		out := make(map[string]interface{}, len(m))
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
