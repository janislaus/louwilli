package service

import (
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"louie-web-administrator/louie_kafka"
	"louie-web-administrator/repository"
	"louie-web-administrator/websocket"
	"strings"
)

type GameStateChecker struct {
	UserService         UserService
	GameService         GameService
	GameDashboardSocket websocket.GameDashboardSocket
	AdminUiSocket       websocket.AdminUiWebsocket
}

func (changer *GameStateChecker) RunGameStateChecker(
	kafkaMessageChannel chan kafka.Message,
) {
	go changer.receiveFromKafkaChannel(kafkaMessageChannel)
}
func (changer *GameStateChecker) receiveFromKafkaChannel(kafkaMessageChannel chan kafka.Message) {

	for message := range kafkaMessageChannel {
		changer.checkAndUpdateGameState(message.Value)
	}
}
func (changer *GameStateChecker) checkAndUpdateGameState(message []byte) bool {

	eventSuccessfulProcessed := false
	var tmpReceivedEvent louie_kafka.DefaultEvent

	_ = json.Unmarshal(message, &tmpReceivedEvent)

	switch tmpReceivedEvent.Event {

	case louie_kafka.PlayersCanBeReceived:
		eventSuccessfulProcessed = changer.playersCanBeReceivedEvent()

	case louie_kafka.PlayersConfirm:
		eventSuccessfulProcessed = changer.playersConfirmedEvent()

	case louie_kafka.GameDone:
		eventSuccessfulProcessed = changer.gameDoneEvent(message)

	case louie_kafka.CoinDrop:
		eventSuccessfulProcessed = changer.coinDropEvent(message)
	}

	return eventSuccessfulProcessed
}

func (changer *GameStateChecker) playersCanBeReceivedEvent() bool {

	var updatedGame *GameEntry

	currentGame, _ := changer.GameService.GetCurrentGame()

	if currentGame == nil {
		return false
	}

	if ok := changer.playersCanBeReceived(currentGame); ok {
		updatedGame, _ = changer.GameService.UpdateGameState(currentGame.Id, repository.GameReady)
	} else {
		return false
	}

	if updatedGame != nil {
		changer.GameService.SendPlayerReadyMessageToKafka(
			[]louie_kafka.PlayerDisplayName{
				{DisplayName: currentGame.Player1},
				{DisplayName: currentGame.Player2},
				{DisplayName: currentGame.Player3},
			})

		changer.AdminUiSocket.SendToAdminUi(toAdminUiEvent(updatedGame))
	} else {
		return false
	}

	return true
}

func (changer *GameStateChecker) playersCanBeReceived(currentGame *GameEntry) bool {

	if currentGame.State != repository.GameAnnounced {
		log.Printf("get \"PLAYERS_CAN_BE_RECEIVED\" from Louie. Current game state is not \"announced\" ignore\n")
		return false
	}
	log.Printf("get \"PLAYERS_CAN_BE_RECEIVED\" from Louie. Update state and send players to louie\n")

	return true
}

func (changer *GameStateChecker) playersConfirmedEvent() bool {

	var updatedGame *GameEntry

	currentGame, _ := changer.GameService.GetCurrentGame()

	if currentGame == nil {
		return false
	}

	if ok := changer.playersConfirmed(currentGame); ok {

		ranking, _ := changer.GameService.GetRankingsSorted()
		dashboardRanking := ToDashboardRanking(ranking)
		updatedGame, _ = changer.GameService.UpdateGameState(currentGame.Id, repository.GameActive)

		if updatedGame != nil {
			changer.GameDashboardSocket.SendToDashboard(
				&websocket.DashboardSignal{
					DashboardGame:    ToDashboardGameFromGameEntry(updatedGame),
					DashboardRanking: dashboardRanking,
				},
			)
		}

		changer.AdminUiSocket.SendToAdminUi(toAdminUiEvent(updatedGame))
	} else {
		return false
	}

	return true
}
func (changer *GameStateChecker) playersConfirmed(currentGame *GameEntry) bool {

	if currentGame.State != repository.GameReady {
		log.Printf("get \"PLAYERS_CONFIRMED\" from Louie. Current game state is not \"ready\" ignore\n")
		return false
	}
	log.Printf("get \"PLAYERS_CONFIRMED\" from Louie. Update state and send game to dashboard\n")

	return true
}

func (changer *GameStateChecker) gameDoneEvent(message []byte) bool {

	var updatedGame *GameEntry

	currentGame, _ := changer.GameService.GetCurrentGame()

	if currentGame == nil {
		return false
	}

	if ok := changer.gameDone(currentGame, message); ok {
		updatedGame, _ = changer.GameService.UpdateGameState(currentGame.Id, repository.GameFinished)

		if updatedGame != nil {
			gameDoneEvent := changer.unmarshalGameDoneEvent(message)
			currentGameId, _ := changer.parseGameId(currentGame.Id)

			changer.updatePlayerStatistic(*currentGameId, gameDoneEvent)
			changer.GameService.UpdateGameDuration(currentGame.Id, gameDoneEvent.Duration)

			ranking, _ := changer.GameService.GetRankingsSorted()
			dashboardRanking := ToDashboardRanking(ranking)

			changer.GameDashboardSocket.SendToDashboard(
				&websocket.DashboardSignal{
					DashboardGame:    ToDashboardGameFromGameEntry(updatedGame),
					DashboardRanking: dashboardRanking,
				})
			changer.AdminUiSocket.SendToAdminUi(toAdminUiEvent(updatedGame))
		}
	} else {
		return false
	}

	return true
}

func (changer *GameStateChecker) gameDone(currentGame *GameEntry, message []byte) bool {

	changer.unmarshalGameDoneEvent(message)

	if currentGame.State != repository.GameActive {
		log.Printf("get \"GAME_DONE\" from Louie. Current game state is not \"active\" ignore\n")
		return false
	}
	log.Printf("get \"GAME_DONE\" from Louie. Update state and player statistics\n")

	return true
}

func (changer *GameStateChecker) coinDropEvent(message []byte) bool {

	var updatedGame *GameEntry

	ranking, _ := changer.GameService.GetRankingsSorted()
	dashboardRanking := ToDashboardRanking(ranking)

	currentGame, _ := changer.GameService.GetCurrentGame()

	if currentGame == nil {
		return false
	}

	if ok := changer.coinDrop(currentGame); ok {
		coinDropEvent := changer.unmarshalCoinDropEvent(message)
		if ok := changer.GameService.UpdateCoins(coinDropEvent.Name, coinDropEvent.Coins); ok {
			game, _ := changer.GameService.GetCurrentGame()
			updatedGame = game
			changer.GameDashboardSocket.SendToDashboard(
				&websocket.DashboardSignal{
					DashboardGame:    ToDashboardGameFromGameEntry(updatedGame),
					DashboardRanking: dashboardRanking,
				})
		}
	}

	if updatedGame != nil {
		changer.AdminUiSocket.SendToAdminUi(toAdminUiEvent(updatedGame))
	}

	return true
}
func (changer *GameStateChecker) coinDrop(currentGame *GameEntry) bool {

	if currentGame.State != repository.GameActive {
		log.Printf("get \"COIN_DROP\" from Louie. Current game state is not \"active\" ignore\n")
		return false
	}
	log.Printf("get \"COIN_DROP\" from Louie. Update player coins\n")

	return true
}

func (changer *GameStateChecker) parseGameId(gameId string) (*primitive.ObjectID, error) {
	currentGameId, err := primitive.ObjectIDFromHex(gameId)

	if err != nil {
		log.Printf("can not parse current game id: %s\n", err)
		return nil, err
	}

	return &currentGameId, nil
}

func (changer *GameStateChecker) unmarshalGameDoneEvent(message []byte) *louie_kafka.GameDoneEvent {
	var gameDoneEvent *louie_kafka.GameDoneEvent

	err := json.Unmarshal(message, &gameDoneEvent)

	if err != nil {
		log.Printf("failures during unmarshal game done event: %s\n", err)
	}

	return gameDoneEvent
}

func (changer *GameStateChecker) unmarshalCoinDropEvent(message []byte) *louie_kafka.CoinDropEvent {
	var coinDropEvent *louie_kafka.CoinDropEvent

	err := json.Unmarshal(message, &coinDropEvent)

	if err != nil {
		log.Printf("failures during unmarshal coin drop event: %s\n", err)
	}

	return coinDropEvent
}

func (changer *GameStateChecker) updatePlayerStatistic(currentGameId primitive.ObjectID, gameDoneEvent *louie_kafka.GameDoneEvent) {
	users, _ := changer.UserService.GetByGameId(currentGameId)

	for _, user := range users {

		if user.BestDuration > gameDoneEvent.Duration || user.BestDuration == repository.InitialUserDuration {
			user.BestDuration = gameDoneEvent.Duration
		}

		if strings.ToLower(user.DisplayName) == strings.ToLower(gameDoneEvent.WinningPlayer.Name) {
			user.GamesWon += 1
		}

		user.PlayedGames += 1

		_, err := changer.UserService.UpdateStatistic(user)
		if err != nil {
			log.Printf("updating user statistic failed: %s\n", err)
		}
	}
}

func toAdminUiEvent(game *GameEntry) *websocket.AdminUiEvent {

	adminUiEventType, err := websocket.ValueOf(game.State.String())

	if err != nil {
		return nil
	}

	return &websocket.AdminUiEvent{
		EventType:    adminUiEventType,
		KiCoins:      game.KiCoins,
		Player1Coins: game.Player1Coins,
		Player2Coins: game.Player2Coins,
		Player3Coins: game.Player3Coins,
	}
}
