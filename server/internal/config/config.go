package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config struct describes a config entity
type Config struct {
	Debug           bool   `envconfig:"DEBUG" default:"true"`
	ServerAddr      string `envconfig:"SERVER_ADDR" default:"192.168.1.69:8080"`
	RedisServerAddr string `envconfig:"REDIS_ADDR" default:"localhost:6379"`
}

// New is a constructor for server's config
func New() (*Config, error) {
	var config Config
	if err := envconfig.Process("VPN_SERVER", &config); err != nil {
		return nil, err
	}
	return &config, nil
}
