package config

import (
	"log"

	"github.com/spf13/viper"
)

func NewViper(envName string) *viper.Viper {
	v := viper.New()

	v.SetConfigName(envName)
	v.AddConfigPath(".")
	v.SetConfigType("env")

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("cannot parse configuration", err)
	}
	return v
}
