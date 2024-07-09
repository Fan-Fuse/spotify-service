package main

import (
	"context"
	"encoding/json"
	"os"

	"go.uber.org/zap"

	"github.com/Fan-Fuse/spotify-service/clients"
	"github.com/Fan-Fuse/spotify-service/service"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		zap.S().Fatalf("%s: %s", msg, err)
	}
}

type UserMessage struct {
	ID string
}

func init() {
	// Initialize logger
	logger := zap.Must(zap.NewProduction())
	if os.Getenv("APP_ENV") == "development" {
		logger = zap.Must(zap.NewDevelopment())
	}

	zap.ReplaceGlobals(logger)

	// Initialize service clients
	clients.InitConfig(os.Getenv("CONFIG_ADDRESS"))
	clients.InitUserClient(os.Getenv("USER_ADDRESS"))
	clients.InitArtistClient(os.Getenv("ARTIST_ADDRESS"))
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	userQueue, err := ch.QueueDeclare(
		"spotify-user", // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare user queue")

	userMsgs, err := ch.Consume(
		userQueue.Name, // queue
		"",             // consumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	failOnError(err, "Failed to register user consumer")

	artistQueue, err := ch.QueueDeclare(
		"spotify-artist", // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare artist queue")

	artistMsgs, err := ch.Consume(
		artistQueue.Name, // queue
		"",               // consumer
		true,             // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
	failOnError(err, "Failed to register artist consumer")

	var forever chan struct{}

	go func() {
		for d := range userMsgs {
			// Unmarshal the message
			var userMessage UserMessage
			err := json.Unmarshal(d.Body, &userMessage)
			if err != nil {
				zap.S().Errorf("Failed to unmarshal user message: %s", err)
				continue
			}

			zap.S().Infof("Received a spotify-user message: %s", d.Body)

			// Handle the user
			err = service.HandleSpotifyUser(context.Background(), userMessage.ID)
			if err != nil {
				continue
			}
		}
	}()

	go func() {
		for d := range artistMsgs {
			zap.S().Infof("Received a spotify-artist message: %s", d.Body)
		}
	}()

	zap.S().Info(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
