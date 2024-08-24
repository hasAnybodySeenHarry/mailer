package main

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"harry2an.com/mailer/internal/mailer"
)

type application struct {
	mailer *mailer.Mailer
	msgQ   msgQ
	logger *log.Logger
}

type msgQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func main() {
	var cfg config
	loadConfig(&cfg)

	l := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	mailer := mailer.New(cfg.smtp.domain, cfg.smtp.apiKey)

	app := application{
		logger: l,
		mailer: mailer,
	}

	if err := app.connect(cfg.amqp); err != nil {
		app.logger.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	defer app.msgQ.ch.Close()
	defer app.msgQ.conn.Close()

	app.consume()
}
