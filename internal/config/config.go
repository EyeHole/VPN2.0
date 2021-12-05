package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config struct describes a config entity
type Config struct {
	Debug      bool   `envconfig:"DEBUG" default:"true"`
	ServerAddr string `envconfig:"SERVER_ADDR" default:""`
	ServerPort string `envconfig:"SERVER_PORT" default:"8080"`
}

// New is a constructor for server's config
func New() (*Config, error) {
	var config Config
	if err := envconfig.Process("VPN_CLIENT", &config); err != nil {
		return nil, err
	}
	return &config, nil
}
