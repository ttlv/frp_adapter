package config

import (
	"github.com/jinzhu/configor"
)

type Config struct {
	Port string
}

var _config *Config

func MustGetConfig() Config {
	if _config != nil {
		return *_config
	}

	_config = &Config{}
	err := configor.New(&configor.Config{ENVPrefix: "FRP_ADAPTER"}).Load(_config)
	if err != nil {
		panic(err)
	}

	return *_config
}
