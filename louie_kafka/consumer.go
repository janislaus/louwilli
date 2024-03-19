package louie_kafka

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"github.com/thoas/go-funk"
	"log"
)

type ConsumerConfig struct {
	Quit                chan bool
	GameEvents          chan kafka.Message
	TechnicalEvents     chan kafka.Message
	ServerAddress       string
	Topic               string
	TechnicalEventTypes []EventType
}

func (consumerConfig *ConsumerConfig) StartConsumer(ctx context.Context) {

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{consumerConfig.ServerAddress},
		Topic:     consumerConfig.Topic,
		Partition: 0,
		MaxBytes:  10e6, // 10MB
	})
	err := r.SetOffset(kafka.LastOffset)

	if err != nil {
		log.Fatal("defining last offset of kafka reader failed:", err)
	}

	go func() {
		for {
			select {
			case <-consumerConfig.Quit:
				if err := r.Close(); err != nil {
					log.Fatal("failed to close reader:", err)
				}
				log.Print("terminate kafka consumer")
				return // terminate go function
			}
		}
	}()

	for {
		select {
		case <-consumerConfig.Quit:
			return // terminate go function
		default: // otherwise read messages
			m, err := r.ReadMessage(ctx)
			if err != nil {
				log.Fatal("reading messages failed:", err)
				return
			}
			log.Printf("consumed kafka message: %s:%s", m.Key, m.Value)

			var tmpReceivedEvent DefaultEvent
			_ = json.Unmarshal(m.Value, &tmpReceivedEvent)

			if funk.Contains(getTechnicalEventTypes(), tmpReceivedEvent.Event) {
				consumerConfig.TechnicalEvents <- m
			} else if funk.Contains(getGameEventTypes(), tmpReceivedEvent.Event) {
				consumerConfig.GameEvents <- m
			} else {
				log.Printf("can not assign a technical or game event")
			}
		}
	}
}
