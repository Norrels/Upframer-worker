package main

import (
	"log"
	"net/http"
	"os"
	"upframer-worker/internal/application/usecases"
	customerrors "upframer-worker/internal/domain/errors"
	"upframer-worker/internal/domain/ports"
	"upframer-worker/internal/infra/ffmpeg"
	"upframer-worker/internal/infra/rabbit"
	"upframer-worker/internal/infra/storage"

	"github.com/joho/godotenv"
)

const (
	maxRetries = 3
)

func init() {
	rabbit.NewRabbitMQConnection()
}

func main() {
	defer rabbit.RabbitMQClient.CloseConnection()
	queueName := "job-creation"

	if err := rabbit.RabbitMQClient.SetupDLQ(queueName); err != nil {
		log.Fatalf("Failed to setup DLQ: %v", err)
	}

	msgs, err := rabbit.RabbitMQClient.ConsumeRabbitMQQueue(queueName)

	if err != nil {
		log.Fatal(err)
	}

	_ = godotenv.Load()

	bucket := os.Getenv("AWS_BUCKET")
	region := os.Getenv("AWS_REGION")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")
	environment := os.Getenv("ENVIRONMENT")

	var storageAdapter ports.Storage

	if environment == "production" {
		if bucket == "" || region == "" || accessKey == "" || secretKey == "" {
			log.Fatal("FATAL: S3 credentials are required in production environment. Set AWS_BUCKET, AWS_REGION, AWS_ACCESS_KEY_ID, and AWS_SECRET_ACCESS_KEY")
		}

		s3Storage, err := storage.NewS3Storage(bucket, region, accessKey, secretKey, sessionToken)
		if err != nil {
			log.Fatalf("FATAL: Failed to initialize S3 storage in production: %v", err)
		}
		storageAdapter = s3Storage
		log.Println("Using S3 storage (production mode)")
	} else {
		if bucket != "" && region != "" && accessKey != "" && secretKey != "" {
			s3Storage, err := storage.NewS3Storage(bucket, region, accessKey, secretKey, sessionToken)
			if err != nil {
				log.Printf("Failed to initialize S3 storage: %v. Using local storage as fallback.", err)
				storageAdapter = storage.NewLocalStorage("./output")
			} else {
				storageAdapter = s3Storage
				log.Println("Using S3 storage (development mode)")
			}
		} else {
			storageAdapter = storage.NewLocalStorage("./output")
			log.Println("Using local storage (development mode - S3 credentials not provided)")
		}
	}

	processor := ffmpeg.NewFFmpegProcessor(storageAdapter)
	publisher := rabbit.NewRabbitPublisher(rabbit.RabbitMQClient)
	processVideoUseCase := usecases.NewProcessVideoUseCase(processor, publisher)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	go func() {
		port := os.Getenv("HEALTH_CHECK_PORT")
		if port == "" {
			port = "3334"
		}
		log.Printf("Health check server starting on port %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("Health check server error: %v", err)
		}
	}()

	forever := make(chan bool)
	go func() {
		for msg := range msgs {
			log.Println("New message received, processing...")

			retryCount := int32(0)
			if msg.Headers != nil {
				if count, ok := msg.Headers["x-retry-count"].(int32); ok {
					retryCount = count
				}
			}

			err := processVideoUseCase.Execute(msg.Body)

			if err != nil {
				log.Printf("Error when processing: %v", err)

				if customerrors.IsPermanentError(err) {
					log.Printf("Permanent error detected. Sending to DLQ without retry.")
					if dlqErr := publisher.PublishToDLQ(queueName, msg.Body, err.Error(), retryCount); dlqErr != nil {
						log.Printf("Failed to send to DLQ: %v", dlqErr)
					}
					msg.Nack(false, false)
					continue
				}

				if retryCount >= maxRetries {
					log.Printf("Max retries (%d) exceeded. Sending to DLQ.", maxRetries)
					if dlqErr := publisher.PublishToDLQ(queueName, msg.Body, err.Error(), retryCount); dlqErr != nil {
						log.Printf("Failed to send to DLQ: %v", dlqErr)
					}
					msg.Nack(false, false) 
					continue
				}

				log.Printf("Temporary error. Retry %d/%d", retryCount+1, maxRetries)
				if requeueErr := rabbit.RabbitMQClient.RequeuWithRetryCount(queueName, msg.Body, retryCount); requeueErr != nil {
					log.Printf("Failed to requeue message: %v", requeueErr)
				}
				msg.Nack(false, false)
			} else {
				msg.Ack(false)
				log.Println("Successfully processed!")
			}
		}
	}()

	<-forever
}
