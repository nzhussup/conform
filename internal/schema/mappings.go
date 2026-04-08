package schema

import (
	"fmt"

	"github.com/nzhussup/konform/internal/errs"
)

func ValidateStrictMappings(sc *Schema) error {
	if sc == nil {
		return errs.InvalidSchemaNil
	}

	keyOwners := make(map[string]string)
	envOwners := make(map[string]string)

	for _, field := range sc.Fields {
		if field.KeyName != "" {
			if prevOwner, exists := keyOwners[field.KeyName]; exists {
				return fmt.Errorf("%w: conflicting key mapping %q between fields %q and %q", errs.InvalidSchema, field.KeyName, prevOwner, field.Path)
			}
			keyOwners[field.KeyName] = field.Path
		}

		if field.EnvName != "" {
			if prevOwner, exists := envOwners[field.EnvName]; exists {
				return fmt.Errorf("%w: conflicting env mapping %q between fields %q and %q", errs.InvalidSchema, field.EnvName, prevOwner, field.Path)
			}
			envOwners[field.EnvName] = field.Path
		}
	}

	return nil
}
