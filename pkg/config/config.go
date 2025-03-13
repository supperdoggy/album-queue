package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	DatabaseURL  string `envconfig:"DATABASE_URL" required:"true"`
	DatabaseName string `envconfig:"DATABASE_NAME" required:"true"`

	BotToken     string  `envconfig:"BOT_TOKEN" required:"true"`
	BotWhitelist []int64 `envconfig:"BOT_WHITELIST" required:"true"`
}

func NewConfig() (*Config, error) {
	cfg := new(Config)
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
