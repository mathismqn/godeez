package config

type Config struct {
	LicenseToken string `mapstructure:"license_token"`
	SecretKey    string `mapstructure:"secret_key"`
	IV           string `mapstructure:"iv"`
}

var Cfg Config
