package env

import (
	"errors"
	"fmt"
	"os"

	"github.com/nzhussup/konform/internal/decode"
	"github.com/nzhussup/konform/internal/errs"
	"github.com/nzhussup/konform/internal/schema"
)

func Load(sc *schema.Schema) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	fieldErrors := make([]error, 0)

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
			ctx := fmt.Sprintf("env %q -> %s", envName, field.Path)
			fieldErrors = append(fieldErrors, errs.WrapDecode(errs.Decode, ctx, err))
		}
	}

	if len(fieldErrors) > 0 {
		return errors.Join(fieldErrors...)
	}

	return nil
}
