package main

import (
	"fmt"
	"log"

	"github.com/nzhussup/konform"
)

type Config struct {
	AppName     string `validate:"required,min=3,max=100"`
	Version     string `validate:"required"`
	Description string `validate:"required,maxlen=10"`
	Database    struct {
		Host     string `validate:"required"`
		Port     int    `validate:"required,min=1,max=65535,len=5"`
		Username string `validate:"required,minlen=3,maxlen=50"`
		Password string `validate:"required,minlen=15"`
		Name     string `validate:"required,regex=^db"`
	}
	Features struct {
		EnableLogging bool
		EnableMetrics bool
	} `validate:"required"`
	Logging struct {
		Level string `validate:"required,oneof=debug|warn|error"`
	} `validate:"required"`
	NumberOfWorkers int    `validate:"oneof=1|2|4|8"`
	NumberOfRetries int    `validate:"nonzero"`
	URL             string `validate:"required,url"`
	InvalidURL      string `validate:"url"`
	ValidEmail      string `validate:"required,email"`
	InvalidEmail    string `validate:"email"`
}

func main() {
	cfg := Config{}

	if _, err := konform.Load(&cfg, konform.FromJSONFile("config.json")); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n%+v\n", cfg)
}
