package main

import (
	"fmt"
	"log"

	"github.com/nzhussup/conform"
)

// This example demonstrates source precedence:
// YAML is loaded first, then ENV overrides matching fields.
type Config struct {
	Server struct {
		Port int    `default:"8080"`
		Host string `default:"0.0.0.0"`
	} `conf:"server"`

	Database struct {
		URL string `conf:"url" env:"DATABASE_URL" required:"true"`
	} `conf:"database"`

	LogLevel string `conf:"log_level" default:"info"`
}

func main() {
	var cfg Config

	err := conform.Load(&cfg,
		conform.FromYAMLFile("config.yaml"),
		conform.FromEnv(),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", cfg)
}
