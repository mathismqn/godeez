package config

type Config struct {
	ArlCookie string `mapstructure:"arl_cookie"`
	SecretKey string `mapstructure:"secret_key"`
	IV        string `mapstructure:"iv"`
}

var Cfg Config
