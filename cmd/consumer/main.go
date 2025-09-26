package main

import (
	"fmt"
	"log"
	"upframer-worker/internal/application/usecases"
	"upframer-worker/internal/infra/ffmpeg"
	"upframer-worker/internal/infra/rabbit"
	"upframer-worker/internal/infra/storage"
)

func init() {
	rabbit.NewRabbitMQConnection()
}

func main() {
	defer rabbit.RabbitMQClient.CloseConnection()
	queueName := "job-creation"

	msgs, err := rabbit.RabbitMQClient.ConsumeRabbitMQQueue(queueName)

	if err != nil {
		log.Fatal(err)
	}

	localStorage := storage.NewLocalStorage("./output")

	processor := ffmpeg.NewFFmpegProcessor(localStorage)
	publisher := rabbit.NewRabbitPublisher(rabbit.RabbitMQClient)
	processVideoUseCase := usecases.NewProcessVideoUseCase(processor, publisher)

	forever := make(chan bool)
	go func() {
		for msg := range msgs {
			fmt.Println("New message received, processing...")
			err := processVideoUseCase.Execute(msg.Body)

			if err != nil {
				msg.Nack(false, true)
				fmt.Printf("Error when processing: %v", err)
			} else {
				msg.Ack(false)
				fmt.Println("Successfully processed!")
			}
		}
	}()

	<-forever
}
