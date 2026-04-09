package validators

import (
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strings"

	"github.com/nzhussup/konform/internal/errs"
	schematypes "github.com/nzhussup/konform/internal/schema/types"
	"github.com/nzhussup/konform/internal/validate/types"
)

const URLRuleName = "url"

var AllowedURLSchemes = []string{"http", "https"}

func URL(f schematypes.Field, validations *[]types.ValidationResult) {
	if f.Type.Kind() != reflect.String {
		*validations = append(*validations, types.ValidationResult{
			Field: f,
			Err:   fmt.Errorf("%w: %s validation supports only string values", errs.ValidationNonString, URLRuleName),
		})
		return
	}

	parsedURL, err := url.ParseRequestURI(strings.TrimSpace(f.Value.String()))
	if err != nil {
		*validations = append(*validations, types.ValidationResult{
			Field: f,
			Err:   errs.ValidationURL,
		})
		return
	}

	if parsedURL.Scheme == "" || !slices.Contains(AllowedURLSchemes, parsedURL.Scheme) || parsedURL.Host == "" {
		*validations = append(*validations, types.ValidationResult{
			Field: f,
			Err:   fmt.Errorf("%w: expected an absolute URL with one of the allowed schemes %v and a host", errs.ValidationURL, AllowedURLSchemes),
		})
	}
}
