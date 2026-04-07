package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nzhussup/konform"
)

// This example demonstrates JSON loading with strict typed decoding.
//
// Capabilities shown:
// - nested object mapping using dot paths (for flat structs)
// - alias mapping via `key` tags (Cache uses `cache`)
// - defaults and required fields
// - safe coercions: string -> int/bool/float/duration
// - list decoding: []string, []int, []time.Duration
// - custom type decoding via encoding.TextUnmarshaler (LogFormat)
//
// Run:
//
//	go run .
//
// Try failure scenarios by editing config.json, for example:
// - set App.Debug to true while changing AppDebug field type to string
// - set Database.Port to "abc"
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
	// Flat mapping into nested JSON paths.
	AppName             string          `key:"App.Name" required:"true"`
	AppDebug            bool            `key:"App.Debug"`
	StartupTimeout      time.Duration   `key:"App.StartupTimeout" default:"10s"`
	AppAllowedOrigins   []string        `key:"App.AllowedOrigins"`
	AppAdminPorts       []int           `key:"App.AdminPorts"`
	DatabasePort        int             `key:"Database.Port" required:"true"`
	DatabaseURI         string          `key:"Database.URI" required:"true"`
	DatabasePoolSize    int             `key:"Database.PoolSize" default:"20"`
	DatabaseMaxLifetime time.Duration   `key:"Database.MaxLifetime" default:"30m"`
	CacheEnabled        bool            `key:"cache.Enabled" default:"true"`
	CachePort           int             `key:"cache.Port" required:"true"`
	CacheURI            string          `key:"cache.URI" required:"true"`
	SamplingRatio       float64         `key:"Observability.SamplingRatio" default:"0.1"`
	RetryBackoff        time.Duration   `key:"Observability.RetryBackoff" default:"250ms"`
	AlertWindows        []time.Duration `key:"Observability.AlertWindows"`
	LogLevel            string          `key:"Log.Level" default:"info"`
	LogFormat           LogFormat       `key:"Log.Format" default:"json"`
	LogOutputs          []LogFormat     `key:"Log.Outputs"`
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
	} `key:"cache" required:"true"`
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
	Cache    Cache `key:"cache" required:"true"`
	Log      Log
	Obs      Observability `key:"Observability"`
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

	if err := konform.Load(&flatCfg, konform.FromJSONFile("config.json")); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Flat config:\n%+v\n", flatCfg)

	var inlineNestedCfg ConfigInlineNested
	if err := konform.Load(&inlineNestedCfg, konform.FromJSONFile("config.json")); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Inline nested config:\n%+v\n", inlineNestedCfg)

	var nestedCfg ConfigNested
	if err := konform.Load(&nestedCfg, konform.FromJSONFile("config.json")); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Nested config:\n%+v\n", nestedCfg)
}
