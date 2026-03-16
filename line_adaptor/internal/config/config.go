package config

import (
	"log"
	"os"
)

type Config struct {
	ChannelAccessToken string
	ChannelSecret      string
	Port               string
}

func Load() Config {
	token := os.Getenv("CHANNEL_ACCESS_TOKEN")
	if token == "" {
		log.Fatal("CHANNEL_ACCESS_TOKEN is required")
	}

	secret := os.Getenv("CHANNEL_SECRET")
	if secret == "" {
		log.Fatal("CHANNEL_SECRET is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		ChannelAccessToken: token,
		ChannelSecret:      secret,
		Port:               port,
	}
}
