package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nzhussup/conform"
)

// This example demonstrates JSON loading with strict typed decoding:
// - nested keys and aliases (cache)
// - defaults and required fields
// - scalar coercion (string->int/bool/float/duration)
// - list decoding ([]string, []int, []time.Duration)
// - custom decoding through encoding.TextUnmarshaler (LogFormat)
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

type ConfigFlat struct {
	AppName             string          `conf:"App.Name" required:"true"`
	AppDebug            bool            `conf:"App.Debug"`
	StartupTimeout      time.Duration   `conf:"App.StartupTimeout" default:"10s"`
	AppAllowedOrigins   []string        `conf:"App.AllowedOrigins"`
	AppAdminPorts       []int           `conf:"App.AdminPorts"`
	DatabasePort        int             `conf:"Database.Port" required:"true"`
	DatabaseURI         string          `conf:"Database.URI" required:"true"`
	DatabasePoolSize    int             `conf:"Database.PoolSize" default:"20"`
	DatabaseMaxLifetime time.Duration   `conf:"Database.MaxLifetime" default:"30m"`
	CacheEnabled        bool            `conf:"cache.Enabled" default:"true"`
	CachePort           int             `conf:"cache.Port" required:"true"`
	CacheURI            string          `conf:"cache.URI" required:"true"`
	SamplingRatio       float64         `conf:"Observability.SamplingRatio" default:"0.1"`
	RetryBackoff        time.Duration   `conf:"Observability.RetryBackoff" default:"250ms"`
	AlertWindows        []time.Duration `conf:"Observability.AlertWindows"`
	LogLevel            string          `conf:"Log.Level" default:"info"`
	LogFormat           LogFormat       `conf:"Log.Format" default:"json"`
	LogOutputs          []LogFormat     `conf:"Log.Outputs"`
}

type ConfigInlineNested struct {
	App struct {
		Name           string
		Debug          bool
		StartupTimeout time.Duration
		AllowedOrigins []string
		AdminPorts     []int
	}
	Database struct {
		Port        int
		URI         string
		PoolSize    int
		MaxLifetime time.Duration
	}
	Cache struct {
		Enabled bool
		Port    int
		URI     string
	} `conf:"cache" required:"true"`
	Observability struct {
		SamplingRatio float64
		RetryBackoff  time.Duration
		AlertWindows  []time.Duration
	}
	Log struct {
		Level   string
		Format  LogFormat
		Outputs []LogFormat
	}
}

type ConfigNested struct {
	App      App
	Database Database
	Cache    Cache `conf:"cache" required:"true"`
	Log      Log
	Obs      Observability `conf:"Observability"`
}

type App struct {
	Name           string
	Debug          bool
	StartupTimeout time.Duration
	AllowedOrigins []string
	AdminPorts     []int
}

type Database struct {
	Port        int
	URI         string
	PoolSize    int
	MaxLifetime time.Duration
}

type Cache struct {
	Enabled bool
	Port    int
	URI     string
}

type Observability struct {
	SamplingRatio float64
	RetryBackoff  time.Duration
	AlertWindows  []time.Duration
}

type Log struct {
	Level   string
	Format  LogFormat
	Outputs []LogFormat
}

func main() {
	var flatCfg ConfigFlat

	if err := conform.Load(&flatCfg, conform.FromJSONFile("config.json")); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Flat config:\n%+v\n", flatCfg)

	var inlineNestedCfg ConfigInlineNested
	if err := conform.Load(&inlineNestedCfg, conform.FromJSONFile("config.json")); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Inline nested config:\n%+v\n", inlineNestedCfg)

	var nestedCfg ConfigNested
	if err := conform.Load(&nestedCfg, conform.FromJSONFile("config.json")); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Nested config:\n%+v\n", nestedCfg)
}
