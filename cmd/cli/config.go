package main

import (
	"errors"

	"github.com/spf13/viper"
)

type Config struct {
	Logger  LoggerConf
	address string `mapstructure:"ADDRESS"`
	port    string `mapstructure:"PORT"`
}

type LoggerConf struct {
	Level string `mapstructure:"LOG_LEVEL"`
}

func NewConfig() Config {
	return Config{}
}

func (config *Config) Init(path string) error {
	if path == "" {
		err := errors.New("void path to config_cli.env")
		return err
	}

	viper.SetDefault("ADDRESS", "127.0.0.1")
	viper.SetDefault("PORT", "8888")

	viper.SetDefault("LOG_LEVEL", "debug")

	viper.AddConfigPath(path)
	viper.SetConfigName("config_cli")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok { //nolint:errorlint
			return err
		}
	}

	config.address = viper.GetString("ADDRESS")
	config.port = viper.GetString("PORT")

	config.Logger.Level = viper.GetString("LOG_LEVEL")

	return nil
}

func (config *Config) GetAddress() string {
	return config.address
}

func (config *Config) GetPort() string {
	return config.port
}
