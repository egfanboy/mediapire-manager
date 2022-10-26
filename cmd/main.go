package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/egfanboy/mediapire-manager/internal/consul"
	_ "github.com/egfanboy/mediapire-manager/internal/health"
	_ "github.com/egfanboy/mediapire-manager/internal/media"
	_ "github.com/egfanboy/mediapire-manager/internal/node"

	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var cleanupFuncs []func()

func addCleanupFunc(fn func()) {
	cleanupFuncs = append(cleanupFuncs, fn)
}

func main() {

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Info().Msg("Initializing Mediapire Manager")

	log.Debug().Msg("starting webserver")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	defer func() {
		signal.Stop(c)
		log.Info().Msg("Running cleanup functions")
		for _, fn := range cleanupFuncs {
			fn()
		}
	}()

	mainRouter := mux.NewRouter()

	mediaManager := app.GetApp()

	err := consul.NewConsulClient()

	if err != nil {
		log.Error().Err(err).Msgf("Failed to connect to consul")
		os.Exit(1)
	}

	err = consul.RegisterService()

	if err != nil {
		log.Error().Err(err).Msgf("Failed to register service to consul")
		os.Exit(1)
	}

	addCleanupFunc(func() { consul.UnregisterService() })

	for _, c := range mediaManager.ControllerRegistry.GetControllers() {
		for _, b := range c.GetApis() {
			b.Build(mainRouter)
		}
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", mediaManager.Config.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      mainRouter, // Pass our instance of gorilla/mux in.
	}

	go func() {
		err := srv.ListenAndServe()

		if err != nil {
			log.Error().Err(err).Msg("")
			os.Exit(1)
		}
	}()

	addCleanupFunc(func() { srv.Close() })

	log.Info().Msg("Mediapire Manager running")

	<-c
}
