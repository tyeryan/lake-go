package config

import (
	"github.com/spf13/viper"
	"lake-go/db"
)

func LoadConfig(path string) (config db.DatabaseConfig, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName("app")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
