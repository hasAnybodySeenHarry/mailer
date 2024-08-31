package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type config struct {
	env      string
	msgProxy msgProxy
	smtp     smtp
}

type msgProxy struct {
	username string
	password string
	port     int
	host     string
}

type smtp struct {
	domain string
	apiKey string
}

func loadConfig(cfg *config) {
	flag.StringVar(&cfg.env, "env", os.Getenv("ENV"), "The name of the environment")

	flag.StringVar(&cfg.msgProxy.username, "amqp-username", os.Getenv("AMQP_USERNAME"), "The username to connect to AMQP messaging proxy")
	flag.StringVar(&cfg.msgProxy.password, "amqp-password", os.Getenv("AMQP_PASSWORD"), "The password to connect to AMQP messaging proxy")
	flag.StringVar(&cfg.msgProxy.host, "amqp-host", os.Getenv("AMQP_HOST"), "The address to connect to AMQP messaging proxy")
	flag.IntVar(&cfg.msgProxy.port, "amqp-port", getEnvInt("AMQP_PORT", 5672), "The port to connect to AMQP messaging proxy")

	flag.StringVar(&cfg.smtp.apiKey, "mail-api-key", os.Getenv("MAIL_API_KEY"), "The API key to connect to the mailing service")
	flag.StringVar(&cfg.smtp.domain, "mail-domain", os.Getenv("MAIL_DOMAIN"), "The domain of the mailing service")

	flag.Parse()
}

func getEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		fmt.Printf("Invalid value for environment variable %s: %s\n", key, valueStr)
		return defaultValue
	}

	return value
}
