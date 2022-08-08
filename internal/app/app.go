package app

import (
	"sync"

	"github.com/egfanboy/mediapire-common/router"
	"github.com/go-redis/redis/v9"
)

type App struct {
	ControllerRegistry *router.ControllerRegistry
	config

	Redis *redis.Client
}

var a *App

var o = sync.Once{}

func initApp() {
	o.Do(func() {
		if a == nil {

			// TODO: add redis info to a config
			rdb := redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "", // no password set
				DB:       0,  // use default DB
			})

			cfg := parseConfig()

			a = &App{ControllerRegistry: router.NewControllerRegistry(), config: cfg, Redis: rdb}
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
