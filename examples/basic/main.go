package main

import (
	"fmt"
	"log"

	"github.com/nzhussup/konform"
)

// This example demonstrates source precedence:
// YAML is loaded first, then ENV overrides matching fields.
type Config struct {
	App struct {
		Name        string
		Version     string
		Description string
		Author      string
		License     string
	}
	Database struct {
		Port        int
		URI         string
		PoolSize    int
		MaxLifetime string
	}
	Redis struct {
		Host string `env:"REDIS_HOST"`
		Port int    `env:"REDIS_PORT"`
	}
}

func main() {
	var cfg Config

	_, err := konform.Load(&cfg,
		konform.FromJSONFile("config.json"),
		konform.FromYAMLFile("config.yaml"),
		konform.FromEnv(),
		konform.Strict(),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", cfg)
}
