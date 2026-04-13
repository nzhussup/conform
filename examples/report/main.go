package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nzhussup/konform"
)

type Config struct {
	Database struct {
		Host     string `key:"database.host" default:"localhost"`
		Port     int    `key:"database.port" default:"5432"`
		Password string `secret:"1" validate:"required,minlen=15"`
	}
	Application struct {
		Name    string `key:"application.name" default:"myapp"`
		Version string `key:"application.version" default:"latest"`
	}
}

func main() {
	cfg := Config{}

	report, err := konform.Load(&cfg,
		konform.FromJSONFile("config.json"),
		konform.FromYAMLFile("config.yaml"),
		konform.WithUnknownKeySuggestionMode(konform.ModeOff),
	)
	if err != nil {
		log.Fatal(err)
	}

	report.Print(os.Stdout)
	fmt.Printf("\n%+v\n", report)
	fmt.Printf("\n\nKonform Version: %s\n", konform.Version)
}
