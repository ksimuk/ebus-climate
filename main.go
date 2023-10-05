package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/ksimuk/ebus-climate/internal/climate"
	"github.com/ksimuk/ebus-climate/internal/ebusd/client"
	"github.com/ksimuk/ebus-climate/internal/web"
)

func main() {
	parser := argparse.NewParser("ebus-climate", "Proxy between ebusd and ha")

	ebusHost := parser.String("", "ehost", &argparse.Options{Required: false, Help: "Ebusd host", Default: "localhost"})
	ebusPort := parser.Int("", "eport", &argparse.Options{Required: false, Help: "Ebusd port", Default: 8888})

	servicePort := parser.Int("", "port", &argparse.Options{Required: false, Help: "Service port", Default: 1080})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	ebusClient := client.New(*ebusHost, *ebusPort)
	climate := climate.New(ebusClient)
	web.Start(*servicePort, climate)
}
