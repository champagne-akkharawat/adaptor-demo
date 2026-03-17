package config

import (
	"fmt"
	"os"
)

type Config struct {
	ChannelSecret      string
	ChannelAccessToken string
	Port               string
	LogDir             string
}

func Load() (*Config, error) {
	secret := os.Getenv("LINE_CHANNEL_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("required env var LINE_CHANNEL_SECRET is not set")
	}

	token := os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("required env var LINE_CHANNEL_ACCESS_TOKEN is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "./logs"
	}

	return &Config{
		ChannelSecret:      secret,
		ChannelAccessToken: token,
		Port:               port,
		LogDir:             logDir,
	}, nil
}
