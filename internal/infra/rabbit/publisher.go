package rabbit

import (
	"encoding/json"
	"fmt"
	"log"
	"upframer-worker/internal/domain/entities"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	client *RabbitMQ
}

func NewRabbitPublisher(client *RabbitMQ) *RabbitPublisher {
	return &RabbitPublisher{
		client: client,
	}
}

func (p *RabbitPublisher) Publish(queueName string, result *entities.ProcessingResult) error {
	_, err := p.client.Channel.QueueDeclare(queueName, true, false, false, false, nil)

	if err != nil {
		return fmt.Errorf("error declaring queue: %v", err)
	}

	notification := map[string]interface{}{
		"outputPath": result.OutputPath,
		"status":     result.Status,
		"jobId":      result.JobId,
	}

	messageJSON, err := json.Marshal(notification)

	if err != nil {
		return fmt.Errorf("erro ao converter mensagem para JSON: %v", err)
	}

	err = p.client.Channel.Publish(
		"",
		queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         messageJSON,
			DeliveryMode: amqp091.Persistent,
		},
	)

	if err != nil {
		log.Fatal("Error while publ", err)
	}

	return nil
}

func (p *RabbitPublisher) PublishToDLQ(queueName string, originalMessage []byte, reason string, retryCount int32) error {
	dlqExchangeName := queueName + ".dlq.exchange"
	dlqRoutingKey := queueName + ".dlq"

	headers := amqp091.Table{
		"x-original-queue": queueName,
		"x-failure-reason": reason,
		"x-retry-count":    retryCount,
	}

	err := p.client.Channel.Publish(
		dlqExchangeName,
		dlqRoutingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         originalMessage,
			DeliveryMode: amqp091.Persistent,
			Headers:      headers,
		},
	)

	if err != nil {
		return fmt.Errorf("error publishing to DLQ: %v", err)
	}

	log.Printf("Message sent to DLQ. Reason: %s, Retry count: %d", reason, retryCount)
	return nil
}
