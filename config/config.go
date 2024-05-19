package config

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	SvrAddress string `mapstructure:"SERVER_ADDRESS"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
}

func LoadConfig(configFileName string) (Config, error) {
	var config Config

	viper.AddConfigPath(".")
	viper.SetConfigName(configFileName)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file, %s", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return config, err
	}

	return config, nil
}
