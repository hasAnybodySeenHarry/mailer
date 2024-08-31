package mailer

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type Mailer struct {
	mg *mailgun.MailgunImpl
}

func New(domain, apiKey string) *Mailer {
	mg := mailgun.NewMailgun(domain, apiKey)
	return &Mailer{mg: mg}
}

func (m *Mailer) Send(name string, recipient string, token string) error {
	msg := m.mg.NewMessage(
		fmt.Sprintf("Admin <mailgun@%s>", m.mg.Domain()),
		fmt.Sprintf("Welcome %s", name),
		fmt.Sprintf("Your activation token is %s", token),
		recipient,
	)

	maxRetries := 5
	baseDelay := time.Second

	var err error

	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

		_, _, err = m.mg.Send(ctx, msg)
		cancel()

		if err == nil {
			return nil
		}

		delay := time.Duration(math.Pow(2, float64(i))) * baseDelay
		jitter := time.Duration(float64(time.Millisecond) * (0.5 + 0.5*rand.Float64()))
		delay += jitter

		log.Printf("Failed to send email: %v. Retrying in %v...\n", err, delay)
		time.Sleep(delay)
	}

	return err
}
