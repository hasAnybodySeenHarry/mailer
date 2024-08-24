package main

import (
	"encoding/json"
)

func (app *application) consume() {
	q, err := app.msgQ.ch.QueueDeclare(
		"email_queue", // queue name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		app.logger.Fatalf("Failed to declare a queue: %v", err)
	}

	msgs, err := app.msgQ.ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		app.logger.Fatalf("Failed to register a consumer: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var msg struct {
				Email    string `json:"email"`
				Name     string `json:"name"`
				Token    string `json:"token"`
				Template string `json:"template"`
			}

			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				app.logger.Printf("Error decoding message body: %v", err)
				continue
			}

			err = app.mailer.Send(msg.Template, msg.Name, msg.Email, msg.Token)
			if err != nil {
				app.logger.Printf("Failed to send email: %v", err)
			} else {
				app.logger.Printf("Email sent to %s", msg.Email)
			}
		}
	}()

	app.logger.Println("Waiting for messages. To exit press CTRL+C")
	<-forever
}
