package amqp

import (
	"errors"
	"time"

	"github.com/streadway/amqp"
)

// Push will push data onto the queue, and wait for a confirm.
// If no confirms are received until within the resendTimeout,
// it continuously re-sends messages until a confirm is received.
// This will block until the server sends a confirm. Errors are
// only returned if the push action itself fails, see UnsafePush.
func (session *Session) Push(key string, data []byte) error {
	if !session.IsReady {
		return errors.New("failed to push push: not connected")
	}
	for {
		err := session.UnsafePush(key, data)
		if err != nil {
			session.logger.Println("Push failed. Retrying...")
			select {
			case <-session.done:
				return errShutdown
			case <-time.After(resendDelay):
			}
			continue
		}
		select {
		case confirm := <-session.notifyConfirm:
			if confirm.Ack {
				return nil
			}
		case <-time.After(resendDelay):
		}
		session.logger.Println("Push didn't confirm. Retrying...")
	}
}

// UnsafePush will push to the queue without checking for
// confirmation. It returns an error if it fails to connect.
// No guarantees are provided for whether the server will
// receive the message.
func (session *Session) UnsafePush(key string, data []byte) error {
	if !session.IsReady {
		return errNotConnected
	}
	return session.channel.Publish(
		session.exchange, // Exchange
		key,              // Routing key
		false,            // Mandatory
		false,            // Immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: 2,
			Body:         data,
		},
	)
}
