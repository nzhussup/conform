package env

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

type DotEnvFileSource struct {
	path      string
	callerDir string
}

func NewDotEnvFileSource(path string, callerDir string) DotEnvFileSource {
	return DotEnvFileSource{path: path, callerDir: callerDir}
}

func (s DotEnvFileSource) LoadFile(sc *schema.Schema) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	data, err := os.ReadFile(resolvePath(s.path, s.callerDir))
	if err != nil {
		return errs.WrapDecode(errs.DecodeSourceRead, fmt.Sprintf(".env file %q", s.path), err)
	}

	values, err := parseDotEnv(data)
	if err != nil {
		return errs.WrapDecode(errs.DecodeSourceParse, fmt.Sprintf(".env file %q", s.path), err)
	}

	return loadWithLookup(sc, func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}, s.path)
}

func parseDotEnv(data []byte) (map[string]string, error) {
	values := make(map[string]string)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if after, ok := strings.CutPrefix(line, "export "); ok {
			line = strings.TrimSpace(after)
		}

		before, after, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: expected KEY=VALUE", lineNo)
		}

		key := strings.TrimSpace(before)
		if key == "" {
			return nil, fmt.Errorf("line %d: key must not be empty", lineNo)
		}

		value, err := parseDotEnvValue(strings.TrimSpace(after), lineNo)
		if err != nil {
			return nil, err
		}
		values[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return values, nil
}

func parseDotEnvValue(raw string, lineNo int) (string, error) {
	if raw == "" {
		return "", nil
	}

	switch raw[0] {
	case '"':
		return parseDoubleQuotedDotEnvValue(raw, lineNo)
	case '\'':
		return parseSingleQuotedDotEnvValue(raw, lineNo)
	default:
		if idx := strings.Index(raw, " #"); idx >= 0 {
			raw = raw[:idx]
		}
		return strings.TrimSpace(raw), nil
	}
}

func parseSingleQuotedDotEnvValue(raw string, lineNo int) (string, error) {
	end := strings.IndexByte(raw[1:], '\'')
	if end < 0 {
		return "", fmt.Errorf("line %d: unterminated quoted value", lineNo)
	}
	end++

	value := raw[1:end]
	if err := validateDotEnvTail(raw[end+1:], lineNo); err != nil {
		return "", err
	}

	return value, nil
}

func parseDoubleQuotedDotEnvValue(raw string, lineNo int) (string, error) {
	end := -1
	escaped := false
	for i := 1; i < len(raw); i++ {
		ch := raw[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '"' {
			end = i
			break
		}
	}
	if end < 0 {
		return "", fmt.Errorf("line %d: unterminated quoted value", lineNo)
	}

	value, err := strconv.Unquote(raw[:end+1])
	if err != nil {
		return "", fmt.Errorf("line %d: invalid quoted value: %w", lineNo, err)
	}
	if err := validateDotEnvTail(raw[end+1:], lineNo); err != nil {
		return "", err
	}

	return value, nil
}

func validateDotEnvTail(tail string, lineNo int) error {
	tail = strings.TrimSpace(tail)
	if tail == "" || strings.HasPrefix(tail, "#") {
		return nil
	}
	return fmt.Errorf("line %d: invalid trailing characters after quoted value", lineNo)
}

func resolvePath(path string, callerDir string) string {
	if filepath.IsAbs(path) || callerDir == "" {
		return path
	}
	return filepath.Join(callerDir, path)
}
