package service

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"louie-web-administrator/louie_kafka"
	"louie-web-administrator/websocket"
)

type TechnicalEventHandler struct {
	KafkaProducer louie_kafka.KafkaProducer
}

func (t *TechnicalEventHandler) SendConfirmedChangeSideEvent() {

	confirmedChangeSideMessage, err := json.Marshal(louie_kafka.DefaultEvent{
		Event: louie_kafka.ConfirmedChangedSide,
	})

	if err != nil {
		log.Printf("can not marshal confirmed changed side kafka message: %s\n", err)
	}

	messages := []kafka.Message{
		{Value: confirmedChangeSideMessage},
	}

	if err != nil {
		log.Printf("sending player can be received message to kafka failed %s\n", err)
	}

	t.KafkaProducer.WriteKafkaMessage(context.Background(), &messages)
}

func RunTechnicalEventHandler(
	kafkaTechnicalEventChannel chan kafka.Message,
	adminUiChannel chan websocket.AdminUiEvent,
	kafkaProducer louie_kafka.KafkaProducer,
) *TechnicalEventHandler {

	go handleTechnicalKafkaEvents(kafkaTechnicalEventChannel, adminUiChannel)

	return &TechnicalEventHandler{kafkaProducer}
}
func handleTechnicalKafkaEvents(kafkaTechnicalEventChannel chan kafka.Message, adminUiChannel chan websocket.AdminUiEvent) {
	for message := range kafkaTechnicalEventChannel {
		var tmpReceivedEvent louie_kafka.DefaultEvent

		_ = json.Unmarshal(message.Value, &tmpReceivedEvent)

		switch tmpReceivedEvent.Event {

		case louie_kafka.PleaseChangeSide:
			adminUiChannel <- websocket.AdminUiEvent{
				EventType: websocket.PlzChangeSide,
			}
		}
	}
}
