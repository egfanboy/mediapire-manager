package app

import (
	"flag"
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	Port   int    `yaml:"port"`
	Scheme string `yaml:"scheme"`
	Consul struct {
		Scheme  string `yaml:"scheme"`
		Port    int    `yaml:"port"`
		Address string `yaml:"address"`
	} `yaml:"consul"`
}

func parseConfig() (config, error) {

	var conf config

	var configPath string
	flag.StringVar(&configPath, "config", "", "optional path to config file")

	flag.Parse()

	if configPath == "" {
		cwd, err := os.Getwd()

		if err != nil {
			return conf, err
		}

		configPath = cwd + "/config.yaml"
	}

	f, err := os.ReadFile(configPath)

	if err != nil {
		return conf, err
	}

	err = yaml.Unmarshal(f, &conf)

	return conf, err
}
