package web

import (
	"fmt"

	"github.com/ksimuk/ebus-climate/internal/climate"
	"github.com/rs/zerolog/log"
)

func Start(port int, climate *climate.Climate) {
	temp, err := climate.ReturnTemp()
	if err != nil {
		fmt.Println(err.Error())
	}
	log.Info().Msgf("Return temp: %f", temp)

	println("Starting web server on port", port)
}
