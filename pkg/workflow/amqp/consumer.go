package amqp

import (
	"log"
)

func (session *Session) Consume(name string, keys []string, handler func(body []byte)) {
	q, err := session.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	FailOnError(err, "Failed to declare a queue")

	for _, s := range keys {
		log.Printf("Binding queue %s to exchange %s with routing key %s", q.Name, session.exchange, s)
		err = session.channel.QueueBind(
			q.Name,           // queue name
			s,                // routing key
			session.exchange, // exchange
			false,
			nil)
		FailOnError(err, "Failed to bind a queue")
	}

	msgs, err := session.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	FailOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			handler(d.Body)
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
