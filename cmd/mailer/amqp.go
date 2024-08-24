package main

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (app *application) connect(uri string) error {
	var err error

	for {
		app.msgQ.conn, err = amqp.Dial(uri)
		if err == nil {
			break
		}

		app.logger.Printf("Failed to connect to RabbitMQ: %s. Retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second)
	}

	app.logger.Printf("Connected to RabbitMQ")

	if err = app.establishChannel(); err != nil {
		return err
	}

	go app.handleConnectionErrors(uri)
	go app.handleChannelErrors()

	return nil
}

func (app *application) establishChannel() error {
	var err error

	if app.msgQ.conn == nil {
		return fmt.Errorf("the connection object is nil")
	}

	for {
		if app.msgQ.conn != nil && !app.msgQ.conn.IsClosed() {
			app.msgQ.ch, err = app.msgQ.conn.Channel()
			if err == nil {
				break
			}
			app.logger.Printf("Failed to create a RabbitMQ channel: %s. Retrying in 5 seconds...", err)
		}

		app.logger.Println("Waiting for an connection to be established")
		time.Sleep(5 * time.Second)
	}

	_, err = app.msgQ.ch.QueueDeclare(
		"email_queue", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return err
	}

	app.logger.Printf("Channel established and queue declared")
	return nil
}

func (app *application) handleConnectionErrors(uri string) {
	for {
		closeErr := <-app.msgQ.conn.NotifyClose(make(chan *amqp.Error))

		if closeErr != nil {
			app.logger.Printf("Connection closed: %s. Attempting to reconnect...", closeErr)

			if err := app.connect(uri); err != nil {
				app.logger.Fatalf("Failed to reconnect to RabbitMQ: %s", err)
			}
		}
	}
}

func (app *application) handleChannelErrors() {
	for {
		closeErr := <-app.msgQ.ch.NotifyClose(make(chan *amqp.Error))

		if closeErr != nil {
			app.logger.Printf("Channel closed: %s. Attempting to reconnect...", closeErr)

			if app.msgQ.conn.IsClosed() {
				app.logger.Println("No established connection exists. Aborting the handling of channel")
				break
			}

			if err := app.establishChannel(); err != nil {
				app.logger.Fatalf("Failed to reconnect RabbitMQ channel: %s", err)
			}
		}
	}
}
