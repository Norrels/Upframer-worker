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

func (r *RabbitMQ) SetupDLQ(queueName string) error {
	dlqName := queueName + ".dlq"
	dlqExchangeName := queueName + ".dlq.exchange"

	err := r.Channel.ExchangeDeclare(
		dlqExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error declaring DLQ exchange: %v", err)
	}

	_, err = r.Channel.QueueDeclare(
		dlqName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error declaring DLQ: %v", err)
	}

	err = r.Channel.QueueBind(
		dlqName,
		dlqName,
		dlqExchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error binding DLQ: %v", err)
	}

	log.Printf("DLQ setup complete: %s", dlqName)
	return nil
}

func (r *RabbitMQ) RequeuWithRetryCount(queueName string, message []byte, retryCount int32) error {
	headers := amqp091.Table{
		"x-retry-count": retryCount + 1,
	}

	err := r.Channel.Publish(
		"",
		queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         message,
			DeliveryMode: amqp091.Persistent,
			Headers:      headers,
		},
	)

	if err != nil {
		return fmt.Errorf("error requeuing message: %v", err)
	}

	return nil
}
