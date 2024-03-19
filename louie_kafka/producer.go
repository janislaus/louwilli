package louie_kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
)

type KafkaProducer struct {
	ServerAddress string
	Topic         string
}

func (p *KafkaProducer) WriteKafkaMessage(ctx context.Context, messages *[]kafka.Message) {

	connection := newKafkaConnection(ctx, p.ServerAddress, p.Topic)

	_, err := connection.WriteMessages(*messages...)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	if err := connection.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}
