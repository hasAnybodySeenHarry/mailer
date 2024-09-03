package main

import (
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (app *application) connect(uri string) error {
	var err error

	retryAttempts := 6
	retryWait := 5 * time.Second

	err = app.retry(retryAttempts, retryWait, func() error {
		app.msgQ.conn, err = amqp.Dial(uri)
		if err != nil {
			app.logger.Printf("Failed to connect to RabbitMQ: %s. Retrying in 5 seconds...", err)
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	app.logger.Printf("Connected to RabbitMQ")

	if err = app.establishChannel(); err != nil {
		return err
	}

	return nil
}

func (app *application) establishChannel() error {
	ok := app.checkAndLock()
	if !ok {
		app.logger.Println("WARNING: concurrent channel establishment detected. Aborting the current call.")
		return nil
	}

	var err error

	if app.msgQ.conn == nil {
		app.setChannelInSync(false)
		return fmt.Errorf("the connection object is nil")
	}

	retryAttempts := 3
	retryWait := 5 * time.Second

	err = app.retry(retryAttempts, retryWait, func() error {
		if app.msgQ.conn == nil || app.msgQ.conn.IsClosed() {
			// we should add an aggressive break out altogether since if
			// the conn is closed, it's either the server is shutdown or
			// there will be a retry for the conn which will then invoke
			// a concurrent channel retry)
			return fmt.Errorf("the connection has not established yet")
		}

		if app.msgQ.ch == nil || app.msgQ.ch.IsClosed() {
			app.msgQ.ch, err = app.msgQ.conn.Channel()
			if err != nil {
				app.logger.Printf("Failed to create a RabbitMQ channel %v. Retrying in 5 seconds...", err)
				return err
			}
		}

		return nil
	})

	if err != nil {
		app.setChannelInSync(false)
		return err
	}

	app.logger.Printf("Channel established and queue declared")

	app.msgQ.mu.Lock()
	defer app.msgQ.mu.Unlock()
	if !app.msgQ.chanInSync {
		panic(errors.New("the channel sync state is broken"))
	}
	app.msgQ.chanInSync = false

	return nil
}

func (app *application) handleConnectionErrors(uri string) {
	closeErr := <-app.msgQ.conn.NotifyClose(make(chan *amqp.Error))

	select {
	case <-app.close:
		app.logger.Println("The server has been signaled to shutdown and won't retry for connections")
		return
	default:
	}

	if closeErr != nil {
		app.logger.Printf("Connection closed: %s. Attempting to reconnect...", closeErr)

		if err := app.connect(uri); err != nil {
			app.logger.Fatalf("Failed to reconnect to RabbitMQ: %s", err)
		}

		go app.handleConnectionErrors(uri)
		go app.handleChannelErrors()
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
