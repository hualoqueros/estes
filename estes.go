package estes

import (
	_ "encoding/json"
	"fmt"
	"log"
	_ "os"
	_ "strings"
	"time"

	"github.com/streadway/amqp"
)

type RMQ struct {
	Exchange   string
	Channel    *amqp.Channel
	Connection *amqp.Connection
}

type Message struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func InitRMQ(connectionString string, exchange string) (*RMQ, error) {
	log.Printf("Initializing RabbitMQ")

	conn, err := amqp.Dial(connectionString)
	failOnError(err, "Failed to connect to RabbitMQ")

	channel, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	err = channel.ExchangeDeclare(
		exchange, // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	rmq := &RMQ{
		Exchange:   exchange,
		Connection: conn,
		Channel:    channel,
	}

	return rmq, err
}

func (rmq *RMQ) PublishEvent(BodyMessage []byte) error {
	return rmq.Channel.Publish(rmq.Exchange, "", false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         BodyMessage,
	})
}

func (rmq *RMQ) ConsumeEvent(queueName string) (<-chan amqp.Delivery, error) {
	queue, err := rmq.Channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive -> set false so Declare will prevent error when queue is already exists
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, fmt.Sprintf("54 : Failed to declare an queue %+v", queue.Name))

	// binding process only
	err = rmq.Channel.QueueBind(
		queue.Name,   // queue name
		"",           // routing key
		rmq.Exchange, // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := rmq.Channel.Consume(
		queue.Name,              // queue
		"", // consumer
		true,                    // auto-ack
		false,                   // exclusive
		false,                   // no-local
		false,                   // no-wait
		nil,                     // args
	)
	failOnError(err, "Failed to register a consumer")

	return msgs, err

}

