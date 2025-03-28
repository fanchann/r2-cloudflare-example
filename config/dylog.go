package config

import (
	"github.com/spf13/viper"
	"github.com/zakirkun/dy"
)

func NewDyLog(v *viper.Viper) *dy.Logger {
	logLvl := v.GetString("LOG_LVL")

	d := dy.New(
		dy.WithPrefix(v.GetString("LOG_PREFIX")),
		dy.WithLevel(parseLvlFromConfig(logLvl)),
		dy.WithColor(true),
	)
	return d
}

func parseLvlFromConfig(lvl string) dy.Level {
	switch lvl {
	case "DEBUG":
		return dy.DebugLevel
	case "INFO":
		return dy.InfoLevel
	case "WARN":
		return dy.WarnLevel
	case "ERROR":
		return dy.ErrorLevel
	default:
		return dy.InfoLevel
	}

}
