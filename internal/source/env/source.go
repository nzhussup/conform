package env

import (
	"fmt"
	"os"

	"github.com/nzhussup/conform/internal/decode"
	"github.com/nzhussup/conform/internal/errs"
	"github.com/nzhussup/conform/internal/schema"
)

func Load(sc *schema.Schema) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	for _, field := range sc.Fields {
		envName := field.EnvName
		if envName == "" {
			continue
		}

		raw, ok := os.LookupEnv(envName)
		if !ok {
			continue
		}

		if err := decode.SetFieldValue(field, raw); err != nil {
			return fmt.Errorf("%w: env %q -> %s: %w", errs.Decode, envName, field.Path, err)
		}
	}

	return nil
}
