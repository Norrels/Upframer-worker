package rabbit

import (
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

func (r *RabbitMQ) ConsumeRabbitMQQueue(queueName string) (<-chan amqp091.Delivery, error) {

	_, err := r.Channel.QueueDeclare(queueName, true, false, false, false, nil)

	if err != nil {
		return nil, fmt.Errorf("error declaring queue: %v", err)
	}

	err = r.Channel.Qos(1, 0, false)
	if err != nil {
		return nil, fmt.Errorf("error configuring QoS: %v", err)
	}

	msgs, err := r.Channel.Consume(queueName, "", false, false, false, false, nil)

	if err != nil {
		return nil, fmt.Errorf("error consuming the queue: %v", err)
	}

	log.Println("Waiting for messages...")
	return msgs, err
}
