package rabbit

import (
	"log"

	"github.com/rabbitmq/amqp091-go"
)

var RabbitMQClient *RabbitMQ

type RabbitMQ struct {
	Conn    *amqp091.Connection
	Channel *amqp091.Channel
}

func NewRabbitMQConnection() {
	conn, err := amqp091.Dial("amqp://admin:admin@localhost:5672/")

	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a RabbitMQ channel: %s", err)
	}

	RabbitMQClient = &RabbitMQ{
		Conn:    conn,
		Channel: ch,
	}
}

func (r *RabbitMQ) CloseConnection() {
	r.Channel.Close()
	r.Conn.Close()
}
