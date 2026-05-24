package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	AppHost string `mapstructure:"HOST"`
	AppPort string `mapstructure:"PORT"`
	DBHost  string `mapstructure:"DB_HOST"`
	DBPort  string `mapstructure:"DB_PORT"`
	DBUser  string `mapstructure:"DB_USER"`
	DBPass  string `mapstructure:"DB_PASSWORD"`
	DBName  string `mapstructure:"DB_NAME"`
}

func NewConfig() (*Config, error) {
	viper.AutomaticEnv()

	viper.BindEnv("PORT")
	viper.BindEnv("HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASSWORD")
	viper.BindEnv("DB_NAME")

	var config Config

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	if err := viper.ReadInConfig(); err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return nil, fmt.Errorf("error reading config file, %s", err)
		}
	}
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config, %v", err)
	}

	return &config, nil
}

func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%s", c.AppHost, c.AppPort)
}

func (c *Config) GetConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser,
		c.DBPass,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}
