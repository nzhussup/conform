package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nzhussup/conform"
)

// This example focuses on environment-variable loading:
// - scalar coercion from string env vars (int/bool/float/duration)
// - defaults
// - required fields
// - custom type decoding via encoding.TextUnmarshaler
//
// Usage:
//
//	set -a; source .env.local; set +a
//	go run .
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
	AppName        string        `env:"APP_NAME" default:"conform-service"`
	Port           int           `env:"PORT" default:"8080"`
	Debug          bool          `env:"DEBUG" default:"false"`
	SamplingRatio  float64       `env:"SAMPLING_RATIO" default:"0.1"`
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT" default:"2s"`
	LogLevel       string        `env:"LOG_LEVEL" default:"info"`
	LogFormat      LogFormat     `env:"LOG_FORMAT" default:"json"`
	DatabaseURL    string        `env:"DATABASE_URL" required:"true"`
}

func main() {
	var cfg Config

	if err := conform.Load(&cfg, conform.FromEnv()); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", cfg)
}
