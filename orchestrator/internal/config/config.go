package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	RabbitMQ struct {
		URL string `yaml:"url"`
	} `yaml:"rabbitmq"`
	Exchange struct {
		Name string `yaml:"name"`
		Type string `yaml:"type"`
	} `yaml:"exchange"`
	Queue struct {
		Name string `yaml:"name"`
		DLX  string `yaml:"dlx"`
	} `yaml:"queue"`
	Bindings []string `yaml:"bindings"`
}

var Cfg Config

func LoadConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &Cfg)
}

func init() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "configs/config.dev.yaml"
	}
	if err := LoadConfig(cfgPath); err != nil {
		panic(err)
	}
}
