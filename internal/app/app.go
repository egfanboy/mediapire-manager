package app

import (
	"os"
	"sync"

	"github.com/egfanboy/mediapire-common/router"

	"github.com/rs/zerolog/log"
)

type App struct {
	ControllerRegistry *router.ControllerRegistry
	Config             config
}

var a *App

var o = sync.Once{}

func initApp() {
	o.Do(func() {
		if a == nil {

			cfg, err := parseConfig()

			if err != nil {
				log.Error().Err(err).Msgf("Failed to read app config file.")
				os.Exit(1)
			}

			a = &App{ControllerRegistry: router.NewControllerRegistry(), Config: cfg}
		}
	})
}

func GetApp() *App {
	initApp()

	return a
}

func init() {
	initApp()
}
