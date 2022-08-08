package app

import (
	"sync"

	"github.com/egfanboy/mediapire-common/router"
)

type App struct {
	ControllerRegistry *router.ControllerRegistry
	config
}

var a *App

var o = sync.Once{}

func initApp() {
	o.Do(func() {
		if a == nil {

			cfg := parseConfig()

			a = &App{ControllerRegistry: router.NewControllerRegistry(), config: cfg}
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
