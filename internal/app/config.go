package app

import (
	"flag"
)

type config struct {
	Port int
}

func parseConfig() config {

	var cfg config

	flag.IntVar(&cfg.Port, "port", 9898, "Server port to listen on")

	flag.Parse()

	return cfg
}
