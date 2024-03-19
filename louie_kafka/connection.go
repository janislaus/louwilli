package louie_kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

func newKafkaConnection(ctx context.Context, serverAddress string, topic string) *kafka.Conn {
	conn, err := kafka.DialLeader(ctx, "tcp", serverAddress, topic, 0)
	if err != nil {
		log.Fatal("failed to create kafka connection:", err)
	}

	err = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	if err != nil {
		log.Fatal("failed to define write deadline for kafka connection:", err)
	}

	return conn
}
