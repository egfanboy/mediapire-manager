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
	"github.com/egfanboy/mediapire-manager/internal/media"
	"github.com/egfanboy/mediapire-manager/internal/node"
	_ "github.com/egfanboy/mediapire-manager/internal/settings"
	_ "github.com/egfanboy/mediapire-manager/internal/transfer"

	// APIs - end

	// messaging - start
	"github.com/egfanboy/mediapire-manager/internal/node/connectivity"
	// messaging - end

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

	nodeService, err := node.NewNodeService()
	if err != nil {
		log.Error().Err(err).Msg("")
		os.Exit(1)
	}

	nodes, err := nodeService.GetAllNodes(ctx)
	if err != nil {
		log.Error().Err(err).Msg("")
		os.Exit(1)
	}

	for _, node := range nodes {
		connectivity.WatchNode(node.Name, node.Id)
	}

	syncService, err := media.NewMediaSyncService(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to create sync service")
		os.Exit(1)
	}

	err = syncService.SyncFromAllNodes(ctx)
	if err != nil {
		// do not exit, just log the error
		log.Error().Err(err).Msg("failed to sync media from all media host nodes")
	}

	if err != nil {
		// do not exit, just log the error
		log.Error().Err(err).Msg("failed to sync media from all media host nodes")
	}

	log.Info().Msg("Mediapire Manager running")

	<-c
}
