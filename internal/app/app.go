package app

import (
	"os"
	"sync"

	"github.com/egfanboy/mediapire-common/router"
	"github.com/google/uuid"

	"github.com/rs/zerolog/log"
)

type App struct {
	ControllerRegistry *router.ControllerRegistry
	Config             config
	NodeId             uuid.UUID
}

var a *App

var o = sync.Once{}

func initApp() {
	if a == nil {

		cfg, err := parseConfig()
		if err != nil {
			log.Error().Err(err).Msgf("Failed to read app config file.")
			os.Exit(1)
		}

		a = &App{ControllerRegistry: router.NewControllerRegistry(), Config: cfg}
	}

	// Create the download path from the config in case it does not exist
	err := os.MkdirAll(a.Config.DownloadPath, os.ModePerm)
	if err != nil {
		return
	}
}

func createApp() {
	o.Do(initApp)
}

func GetApp() *App {
	createApp()

	return a
}

func init() {
	initApp()
}
