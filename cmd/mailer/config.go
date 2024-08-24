package main

import (
	"flag"
	"os"
)

type config struct {
	env  string
	amqp string
	smtp smtp
}

type smtp struct {
	domain string
	apiKey string
}

func loadConfig(cfg *config) {
	flag.StringVar(&cfg.env, "env", os.Getenv("ENV"), "The name of the environment")
	flag.StringVar(&cfg.amqp, "amqp-uri", os.Getenv("AMQP_URI"), "The URI of the AMQP messaging proxy")

	flag.StringVar(&cfg.smtp.apiKey, "mail-api-key", os.Getenv("MAIL_API_KEY"), "The API key to connect to the mailing service")
	flag.StringVar(&cfg.smtp.domain, "mail-domain", os.Getenv("MAIL_DOMAIN"), "The domain of the mailing service")

	flag.Parse()
}
