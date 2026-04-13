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
	return LoadWithPrefix(sc, "")
}

func LoadWithPrefix(sc *schema.Schema, prefix string) error {
	return loadWithLookup(sc, os.LookupEnv, "env", prefix)
}

func loadWithLookup(sc *schema.Schema, lookup func(string) (string, bool), sourcePrefix string, envPrefix string) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	fieldErrors := make([]error, 0)

	for i := range sc.Fields {
		field := sc.Fields[i]
		envName := field.EnvName
		if envName == "" {
			continue
		}
		lookupName := envPrefix + envName

		raw, ok := lookup(lookupName)
		if !ok {
			continue
		}

		if err := decode.SetFieldValue(field, raw); err != nil {
			ctx := fmt.Sprintf("env %q -> %s", lookupName, field.Path)
			fieldErrors = append(fieldErrors, errs.WrapDecode(errs.Decode, ctx, err))
			continue
		}
		sc.Fields[i].Source = sourcePrefix + ":" + lookupName
	}

	if len(fieldErrors) > 0 {
		return errors.Join(fieldErrors...)
	}

	return nil
}
