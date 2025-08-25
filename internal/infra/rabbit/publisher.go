package rabbit

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
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
		"output_path": result.OutputPath,
		"success":     result.Success,
		"error_msg":   result.ErrorMsg,
		"timestamp":   time.Now().Unix(),
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
