package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/egfanboy/mediapire-manager/internal/consul"
	"github.com/egfanboy/mediapire-manager/internal/mongo"
	"github.com/egfanboy/mediapire-manager/internal/rabbitmq"

	// APIs - start

	_ "github.com/egfanboy/mediapire-manager/internal/changeset"
	_ "github.com/egfanboy/mediapire-manager/internal/health"
	_ "github.com/egfanboy/mediapire-manager/internal/media"
	_ "github.com/egfanboy/mediapire-manager/internal/node"
	_ "github.com/egfanboy/mediapire-manager/internal/settings"
	_ "github.com/egfanboy/mediapire-manager/internal/transfer"

	// APIs - end

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var cleanupFuncs []func()

func addCleanupFunc(fn func()) {
	cleanupFuncs = append(cleanupFuncs, fn)
}

func main() {
	ctx := context.Background()

	err := rabbitmq.Setup(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to connect to rabbitmq")
		os.Exit(1)
	}

	addCleanupFunc(func() { rabbitmq.Cleanup() })

	err = mongo.InitMongo(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to connect to mongoDB")
		os.Exit(1)
	}

	addCleanupFunc(func() { mongo.CleanUpMongo(ctx) })

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

	err = consul.NewConsulClient()
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

	mediaManager := app.GetApp()
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
