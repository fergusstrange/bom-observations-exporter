package main

import "github.com/crgimenes/goconfig"

type Config struct {
	ElasticSearchURL string `cfg:"elasticsearch_url" cfgDefault:"http://localhost:9200"`
}

func LoadConfig() (*Config, error) {
	config := new(Config)
	if err := goconfig.Parse(config); err != nil {
		return nil, err
	}
	return config, nil
}
