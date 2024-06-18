package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"os"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// defining the struct of the message which is send from RabbitMQ.
type StockMessage struct {
	Company   string  `json:"company"`
	EventType string  `json:"eventType"`
	Price     float64 `json:"price"`
}

// Reading to enviorment variable.
var companyName string = os.Getenv("COMPANYNAME")
var mongoURL string = os.Getenv("MONGODB_URL")

// main is the entry point of the application.
func main() {

	// Connecting to MongoDB.
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURL))
	failOnError(err, "Failed to connect to MongoDB")

	// Test connection with ping.
	err = client.Ping(context.TODO(), readpref.Primary())
	failOnError(err, "Failed to ping MongoDB")

	// Create new Database with Collection
	usersCollection := client.Database("Stockmarket").Collection("Stocks")

	// Connecting with RabbitMQ.
	conn, err := amqp.Dial("amqp://stockmarket:supersecret123@rabbitmq:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// Opening a channel.
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// If no company name is defiened, we will throw an error.
	if companyName == "" {
		err := errors.New("empty companyname")
		failOnError(err, "Company Name not defiened")
	}

	// Starting stock reader for company.
	log.Printf("Starting stock reader for company: %s", companyName)
	stockReader(ch, usersCollection)

	log.Printf("All stock publishers are running. Press CTRL+C to exit.")
	select {}
}

// failOnError is a helper function to log error messages.
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// aggregateStock avarages the price from a batch.
func aggregateStock(prices []float64) float64 {
	total := 0.0
	for _, price := range prices {
		total += price
	}
	return total / float64(len(prices))
}

// saveToMongoDB saves the average to the MongoDB
func saveToMongoDB(usersCollection *mongo.Collection, average float64) {

	avg := math.Round(average)

	// Defiening the document structur.
	document := bson.M{
		"company":  companyName,
		"avgPrice": avg,
	}

	// If the insertion of the document takes longer as 5 seconds, it will be canceled.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Inserting the document into the MongoDB.
	_, err := usersCollection.InsertOne(ctx, document)
	failOnError(err, "Failed to insert document into MongoDB")
}

// stockReader reads msgs from Queue
func stockReader(ch *amqp.Channel, usersCollection *mongo.Collection) {

	// Declaring the batchsize.
	batchSize := 1000

	// Declating slice for the prices.
	var prices []float64

	// Declaring a queue.
	q, err := ch.QueueDeclare(
		companyName, // Name of the Queue is the same as the stock.
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	failOnError(err, "Failed to declare a queue")

	for {
		// Consume message from Queue.
		msgs, err := ch.Consume(
			q.Name,
			"",
			true, // Auto-Acknowledge -> Delete message when successfully consumed
			false,
			false,
			false,
			nil,
		)
		failOnError(err, "Failed to register a consumer")

		log.Println("Waiting for messages...")

		// Putting the send prices into the slice.
		for i := 0; i < batchSize; i++ {
			msg := <-msgs

			var stockMessage StockMessage
			err := json.Unmarshal(msg.Body, &stockMessage)
			failOnError(err, "Failed to unmarshal message")

			prices = append(prices, stockMessage.Price)
		}

		average := aggregateStock(prices)
		log.Printf("Calculated average price: %f", average)
		saveToMongoDB(usersCollection, average)
	}
}
