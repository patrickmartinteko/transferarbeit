package main

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

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

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	failOnError(err, "Failed to connect to MongoDB")

	usersCollection := client.Database("Stock").Collection(companyName)

	conn, err := amqp.Dial("amqp://stockmarket:supersecret123@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declaring the variable for the queue name, which is the same
	// as the stock company.
	var companyName string = "MSFT"

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
