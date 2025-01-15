package brokers

import (
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/salesof7/eventbus/internal/eventbus"
)

type KafkaEventBroker struct {
	ProducerConfig *kafka.ConfigMap
	ConsumerConfig *kafka.ConfigMap
	Topics         []string
}

func NewKafkaEventBroker(producerConfig, consumerConfig *kafka.ConfigMap, topics []string) *KafkaEventBroker {
	return &KafkaEventBroker{
		ProducerConfig: producerConfig,
		ConsumerConfig: consumerConfig,
		Topics:         topics,
	}
}

func (k *KafkaEventBroker) Publish(eventPayload *eventbus.EventPayload, topic string) error {
	producer, err := kafka.NewProducer(k.ProducerConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	defer producer.Close()

	body, err := json.Marshal(eventPayload)
	if err != nil {
		return err
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          body,
	}

	err = producer.Produce(msg, nil)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	e := <-producer.Events()
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		return fmt.Errorf("delivery failed: %w", m.TopicPartition.Error)
	}
	return nil
}

func (k *KafkaEventBroker) Consume(responseQueue chan *eventbus.EventPayload, errorCallback chan error) {
	consumer, err := kafka.NewConsumer(k.ConsumerConfig)
	if err != nil {
		panic(fmt.Errorf("failed to create Kafka consumer: %w", err))
	}
	defer consumer.Close()

	err = consumer.SubscribeTopics(k.Topics, nil)
	if err != nil {
		panic(fmt.Errorf("failed to subscribe to topics: %w", err))
	}

	for {
		msg, err := consumer.ReadMessage(-1)
		if err == nil {
			eventPayload := &eventbus.EventPayload{}
			err := json.Unmarshal(msg.Value, eventPayload)
			if err != nil {
				errorCallback <- fmt.Errorf("error unmarshalling event payload: %v", err)
			}
			responseQueue <- eventPayload
		} else {
			errorCallback <- fmt.Errorf("failed to read message: %w", err)
		}
	}
}
