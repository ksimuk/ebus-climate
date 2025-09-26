package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/ksimuk/ebus-climate/internal/config"
	"github.com/ksimuk/ebus-climate/internal/web"
)

func main() {
	parser := argparse.NewParser("ebus-climate", "ebus climate service")

	configPath := parser.String("", "config", &argparse.Options{Required: true, Help: "Config Path"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	config, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	web.Start(*config)
}
