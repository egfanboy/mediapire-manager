package main

import (
	"fmt"

	"github.com/egfanboy/mediapire-manager/internal/app"
	_ "github.com/egfanboy/mediapire-manager/internal/media"
	_ "github.com/egfanboy/mediapire-manager/internal/node"

	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Debug().Msg("starting webserver")
	mainRouter := mux.NewRouter()

	mediaManager := app.GetApp()

	for _, c := range mediaManager.ControllerRegistry.GetControllers() {
		for _, b := range c.GetApis() {
			b.Build(mainRouter)
		}
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", mediaManager.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      mainRouter, // Pass our instance of gorilla/mux in.
	}

	srv.ListenAndServe()
}
