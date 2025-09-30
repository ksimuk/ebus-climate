package main

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"os"
	"os/signal"
	"syscall"

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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	server := web.GetServer(*config)

	go func() {
		<-sigs
		log.Info().Msg("Shutting down...")
		server.Shutdown()
		os.Exit(0)
	}()

	server.Start()
}
