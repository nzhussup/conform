# Getting Started

## Install

```bash
go get github.com/nzhussup/konform
```

## Basic usage

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nzhussup/konform"
)

type Config struct {
	Server struct {
		Host string `key:"server.host" default:"127.0.0.1"`
		Port int    `key:"server.port" default:"8080" env:"PORT"`
	}
	Database struct {
		URL string `key:"database.url" env:"DATABASE_URL" validate:"required" secret:"true"`
	}
}

func main() {
	var cfg Config

	report, err := konform.Load(
		&cfg,
		konform.FromYAMLFile("config.yaml"),
		konform.FromEnv(),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", cfg)
	report.Print(os.Stdout)
}
```

## Source order and precedence

`konform` applies values in this order (by call order):

```text
defaults < first source < ... < last source
```

Typical setup:

```go
report, err := konform.Load(
	&cfg,
	konform.FromJSONFile("config.json"),
	konform.FromYAMLFile("config.yaml"),
	konform.FromEnv(),
)
```

In this case, env has highest precedence.
