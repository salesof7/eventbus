package brokers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/salesof7/eventbus/internal/eventbus"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	User              string
	Password          string
	Host              string
	Port              string
	Vhost             string
	ConsumerQueueName string
	ConsumerName      string
	AutoAck           bool
	Args              amqp.Table
	Channel           *amqp.Channel
}

func NewRabbitMQ() *RabbitMQ {
	rabbitMQArgs := amqp.Table{}
	rabbitMQArgs["x-dead-letter-exchange"] = os.Getenv("RABBITMQ_DLX")

	rabbitMQ := RabbitMQ{
		User:              os.Getenv("RABBITMQ_DEFAULT_USER"),
		Password:          os.Getenv("RABBITMQ_DEFAULT_PASS"),
		Host:              os.Getenv("RABBITMQ_DEFAULT_HOST"),
		Port:              os.Getenv("RABBITMQ_DEFAULT_PORT"),
		Vhost:             os.Getenv("RABBITMQ_DEFAULT_VHOST"),
		ConsumerQueueName: os.Getenv("RABBITMQ_CONSUMER_QUEUE_NAME"),
		ConsumerName:      os.Getenv("RABBITMQ_CONSUMER_NAME"),
		AutoAck:           false,
		Args:              rabbitMQArgs,
	}
	return &rabbitMQ
}

func (r *RabbitMQ) Connect() *amqp.Channel {
	dsn := "amqp://" + r.User + ":" + r.Password + "@" + r.Host + ":" + r.Port + r.Vhost
	conn, err := amqp.Dial(dsn)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to RabbitMQ", err)
	}

	r.Channel, err = conn.Channel()
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open a channel", err)
	}
	return r.Channel
}

func (r *RabbitMQ) Publish(eventPayload *eventbus.EventPayload, topic string) error {
	body, err := json.Marshal(eventPayload)
	if err != nil {
		return err
	}

	err = r.Channel.Publish(
		"",    // exchange vazio (ou pode ser configurado conforme necessário)
		topic, // nome da fila como routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	return err
}

func (r *RabbitMQ) Consume(responseQueue chan *eventbus.EventPayload, errorCallback chan error) {
	q, err := r.Channel.QueueDeclare(
		r.ConsumerQueueName, // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		r.Args,              // arguments
	)
	if err != nil {
		panic(err)
	}

	incomingMessages, err := r.Channel.Consume(
		q.Name,         // queue
		r.ConsumerName, // consumer
		r.AutoAck,      // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		panic(err)
	}

	go func() {
		for message := range incomingMessages {
			log.Println("incoming new message")

			eventPayload := &eventbus.EventPayload{}
			err := json.Unmarshal(message.Body, eventPayload)
			if err != nil {
				errorCallback <- fmt.Errorf("error unmarshalling event payload: %v", err)
			}
			responseQueue <- eventPayload
		}
		log.Println("rabbitmq channel closed")
		close(responseQueue)
	}()
}
