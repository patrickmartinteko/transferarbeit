package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	amqp "github.com/rabbitmq/amqp091-go"
)

// failOnError is a helper function to log error messages.
func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// aggregateStock avarages the price from a batch sise and inputs it into the MongoDB:
func aggregateStock(msgsBody json) float64 {

	var sum float64 = 0
	var len float64 = 0

	for nzCount := 0; nzCount <= len; nzCount++ {
		prices := append(prices, msgsBody.price)
		sum += msgsBody.price
	}

	avg := (sum / len)

	entry := bson.D{{"CompanyName", "AvgPrice"}, {companyName, avg}}
	result, err := usersCollection.InsertOne(context.TODO(), entry)
}

func main() {

	// Connecting to MongoDB.
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017,localhost:27018,localhost:27019/?replicaSet=rs0"))
	failOnError(err, "Failed to connect to MongoDB")

	// Test connection with ping
	err = client.Ping(context.TODO(), readpref.Primary())
	failOnError(err, "Failed to ping MongoDB")

	// Creating Database and Collection.
	usersCollection := client.Database("Market").Collection("Stocks")

	// Connecting to RabbitMQ.
	conn, err := amqp.Dial("amqp://stockmarket:supersecret123@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declaring the variable for the queue name, which is the same
	// as the stock company.
	var companyName string = os.Getenv("COMPANYNAME")

	q, err := ch.QueueDeclare(
		companyName, // name of the queue is the same as the stock
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	for d := range msgs {
		var msgsBody json = d.Body
		aggregateStock(msgsBody)
	}
}
