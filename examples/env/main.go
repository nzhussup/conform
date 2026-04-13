package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nzhussup/konform"
)

// This example focuses on environment-variable loading:
// - scalar coercion from string env vars (int/bool/float/duration)
// - defaults
// - required fields
// - custom type decoding via encoding.TextUnmarshaler
// - strict behavior (only safe conversions toward target types)
//
// Usage:
//
//	cat > .env.local <<EOF
//	DATABASE_URL=postgres://localhost:5432/app
//	EOF
//	go run .
//
// Notes:
// - ENV values are always strings, so decoding is string -> typed field.
// - Examples here: "9090" -> int, "true" -> bool, "1500ms" -> duration.
// - Missing required values produce a validation error with field names.
type LogFormat string

func (f *LogFormat) UnmarshalText(text []byte) error {
	v := strings.ToLower(string(text))
	switch v {
	case "json", "text":
		*f = LogFormat(v)
		return nil
	default:
		return fmt.Errorf("invalid log format %q", string(text))
	}
}

type Config struct {
	// defaulted string
	AppName string `env:"APP_NAME" default:"konform-service"`
	// string -> int
	Port int `env:"PORT" default:"8080"`
	// string -> bool
	Debug bool `env:"DEBUG" default:"false"`
	// string -> float64
	SamplingRatio float64 `env:"SAMPLING_RATIO" default:"0.1"`
	// string -> time.Duration
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT" default:"2s"`
	LogLevel       string        `env:"LOG_LEVEL" default:"info"`
	// string -> custom TextUnmarshaler
	LogFormat LogFormat `env:"LOG_FORMAT" default:"json"`
	// required field
	DatabaseURL string `env:"DATABASE_URL" validate:"required"`
}

func main() {
	var cfg Config

	if _, err := konform.Load(&cfg, konform.FromDotEnvFile(".env.local"), konform.FromEnv()); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", cfg)
}
