package main

import (
	"log"
	"net/http"
	"os"
	"upframer-worker/internal/application/usecases"
	"upframer-worker/internal/domain/ports"
	"upframer-worker/internal/infra/ffmpeg"
	"upframer-worker/internal/infra/rabbit"
	"upframer-worker/internal/infra/storage"

	"github.com/joho/godotenv"
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

	err = godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file", "error", err)
	}

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
			err := processVideoUseCase.Execute(msg.Body)

			if err != nil {
				msg.Nack(false, true)
				log.Printf("Error when processing: %v", err)
			} else {
				msg.Ack(false)
				log.Println("Successfully processed!")
			}
		}
	}()

	<-forever
}
