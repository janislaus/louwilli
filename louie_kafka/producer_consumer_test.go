package louie_kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"louie-web-administrator/repository"
	"os"
	"testing"
	"time"
)

func TestProducerAndConsumer(t *testing.T) {

	if os.Getenv("ENABLE_KAFKA_TEST") == "" {
		t.Skip("Skipping kafka tests")
	}

	ctx := context.Background()
	result := make(chan kafka.Message, 1000)
	quit := make(chan bool)

	messages := createLouieEventKafkaMessages()

	consumerConfig := ConsumerConfig{
		Quit:          quit,
		GameEvents:    result,
		ServerAddress: "localhost:9093",
		Topic:         "test42",
	}

	go consumerConfig.StartConsumer(ctx)

	time.Sleep(2 * time.Second)

	producer := KafkaProducer{
		ServerAddress: "localhost:9093",
		Topic:         "test42",
	}

	go producer.WriteKafkaMessage(ctx, &messages)

	time.Sleep(4 * time.Second)

	quit <- true
	close(quit)
	close(result)

	for message := range result {
		fmt.Printf("message at offset %d: %s = %s\n", message.Offset, string(message.Key), string(message.Value))
	}
}

func createLouieEventKafkaMessages() []kafka.Message {
	gameDone, _ := json.Marshal(GameDoneEvent{
		Sender:   "webserver",
		Event:    GameDone,
		Duration: 300,
		WinningPlayer: winningPlayer{
			Name: "Max",
		},
	})

	playersReady, _ := json.Marshal(PlayersReadyEvent{
		Event:     PlayersReady,
		Timestamp: time.Now().Format(repository.GermanDateTimeFormat),
		Players:   []PlayerDisplayName{{DisplayName: "max1"}, {DisplayName: "max2"}},
	})

	confirmedChangedSide, _ := json.Marshal(DefaultEvent{
		Event: ConfirmedChangedSide,
	})

	coinDrop, _ := json.Marshal(CoinDropEvent{
		Sender: "webserver",
		Event:  CoinDrop,
		Name:   "max",
		Coins:  2,
	})

	resetGame, _ := json.Marshal(resetGame{
		Sender: "webserver",
		Event:  ResetGame,
	})

	playersConfirm, _ := json.Marshal(playersConfirm{
		Sender: "webserver",
		Event:  PlayersConfirm,
	})

	playersCanBeReceived, _ := json.Marshal(PlayersCanBeReceivedEvent{
		Event:  PlayersCanBeReceived,
		Sender: "webserver",
	})

	return []kafka.Message{
		{Value: gameDone},
		{Value: playersReady},
		{Value: confirmedChangedSide},
		{Value: coinDrop},
		{Value: resetGame},
		{Value: playersConfirm},
		{Value: playersCanBeReceived},
	}

}
