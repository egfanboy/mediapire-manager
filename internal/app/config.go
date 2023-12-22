package app

import (
	"flag"
	"os"
	"path"

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
	Rabbit struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Port     int    `yaml:"port"`
		Address  string `yaml:"address"`
	} `yaml:"rabbit"`
	MongoURI     string `yaml:"mongoConnectionURI"`
	DownloadPath string `yaml:"-"`
}

func getDownloadPath() (string, error) {
	basePath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(basePath, ".mediapire", "manager", "downloads"), nil
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
	if err != nil {
		return conf, err
	}

	dlPath, err := getDownloadPath()
	if err != nil {
		return conf, err
	}

	conf.DownloadPath = dlPath

	return conf, err
}
