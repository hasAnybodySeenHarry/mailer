package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (app *application) consume() {
	relay := make(chan os.Signal, 1)
	signal.Notify(relay, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	var err error

	go func() {
		for {
			// queue declaration
			_, err = app.msgQ.ch.QueueDeclare(
				"email_queue",
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				app.logger.Printf("Failed to declare a queue: %v. Retrying in 5 seconds...", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// queue consumption
			msgs, err := app.msgQ.ch.Consume(
				"email_queue", // queue
				"",            // consumer
				true,          // auto-ack
				false,         // exclusive
				false,         // no-local
				false,         // no-wait
				nil,           // args
			)
			if err != nil {
				app.logger.Printf("Failed to register a consumer: %v. Retrying in 5 seconds...", err)
				time.Sleep(5 * time.Second)
				continue
			}
			app.logger.Printf("Consuming from the queue named %s", "email_queue")

			time.Sleep(6 * time.Hour)

			// cancellation watch loop
			for {
				// exit or handle tasks
				select {
				case <-ctx.Done():
					return
				case d, ok := <-msgs:
					if !ok {
						app.logger.Println("Error reading message from the channel.")
						break
					}

					app.background(func() {
						err := app.handleMailJob(d)
						if err != nil {
							app.logger.Println(err)
						}
					})
				}
			}
		}
	}()

	app.logger.Println("Ready to fulfill the jobs...")
	<-relay
	cancel()

	app.logger.Println("Waiting for goroutines...")
	app.wg.Wait()

	app.close <- struct{}{} // Signal not to retry the connection anymore

	err = app.msgQ.ch.Close()
	if err != nil {
		app.logger.Printf("Error closing channel: %v", err)
	}

	err = app.msgQ.conn.Close()
	if err != nil {
		app.logger.Printf("Error closing connection: %v", err)
	}

	app.logger.Println("Shutdown completed.")
}

func (app *application) handleMailJob(d amqp.Delivery) error {
	var msg struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Token    string `json:"token"`
		Template string `json:"template"`
	}

	err := json.Unmarshal(d.Body, &msg)
	if err != nil {
		return fmt.Errorf("error decoding message body: %v", err)
	}

	err = app.mailer.Send(msg.Name, msg.Email, msg.Token)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	app.logger.Printf("Email sent to %s", msg.Email)
	return nil
}
