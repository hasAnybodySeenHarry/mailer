package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"harry2an.com/mailer/internal/mailer"
)

type application struct {
	mailer *mailer.Mailer
	msgQ   msgQ
	logger *log.Logger
	wg     sync.WaitGroup
	close  chan struct{}
}

type msgQ struct {
	conn       *amqp.Connection
	ch         *amqp.Channel
	chanInSync bool
	mu         sync.Mutex
}

func main() {
	var cfg config
	loadConfig(&cfg)

	l := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	mailer := mailer.New(cfg.smtp.domain, cfg.smtp.apiKey)

	app := application{
		logger: l,
		mailer: mailer,
		close:  make(chan struct{}, 1),
	}

	uri := fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.msgProxy.username, cfg.msgProxy.password, cfg.msgProxy.host, cfg.msgProxy.port)
	if err := app.connect(uri); err != nil {
		app.logger.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	go app.handleConnectionErrors(uri)

	defer app.msgQ.ch.Close()
	defer app.msgQ.conn.Close()

	app.consume()
}
