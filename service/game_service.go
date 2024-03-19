package service

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"github.com/thoas/go-funk"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"louie-web-administrator/louie_kafka"
	"louie-web-administrator/repository"
	"louie-web-administrator/websocket"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

type GameService interface {
	SendPlayerReadyMessageToKafka(playerDisplayNames []louie_kafka.PlayerDisplayName)
	CreateGame(gameMembers []repository.RegisteredUser) (*primitive.ObjectID, error)
	RemoveGame(gameId string) (*mongo.DeleteResult, error)
	UpdateGameState(gameId string, state repository.GameState) (*GameEntry, error)
	GetRankingsSorted() ([]Ranking, error)
	GetCurrentGame() (*GameEntry, error)
	GetCurrentDashboardState() (*websocket.DashboardSignal, error)
	UpdateGameDuration(gameId string, duration float64)
	UpdateCoins(player string, coins int) bool
}

type GameEntry struct {
	Id           string
	Duration     float64
	KiName       string
	KiCoins      int
	Player1      string
	Player1Coins int
	Player2      string
	Player2Coins int
	Player3      string
	Player3Coins int
	State        repository.GameState
}

type GameSer struct {
	UserRepository repository.UserRepository
	GameRepository repository.GameRepository
	KafkaProducer  louie_kafka.KafkaProducer
}

func (g *GameSer) SendPlayerReadyMessageToKafka(playerDisplayNames []louie_kafka.PlayerDisplayName) {

	playerCanBeReceived, err := json.Marshal(louie_kafka.PlayersReadyEvent{
		Event:     louie_kafka.PlayersReady,
		Players:   playerDisplayNames,
		Timestamp: strconv.FormatInt(time.Now().UnixMilli(), 10),
	})

	messages := []kafka.Message{
		{Value: playerCanBeReceived},
	}

	if err != nil {
		log.Printf("sending player ready message to kafka failed %s\n", err)
	}

	g.KafkaProducer.WriteKafkaMessage(context.Background(), &messages)
}

func (g *GameSer) CreateGame(gameMembers []repository.RegisteredUser) (*primitive.ObjectID, error) {

	gameId, err := g.GameRepository.CreateGame(gameMembers)

	if err != nil {
		return nil, err
	}

	for _, user := range gameMembers {
		_, err := g.UserRepository.UpdateGameRelationship(&user.Id, gameId)
		if err != nil {
			log.Printf("creating of game failed %s\n", err)
			return nil, err
		}
	}

	return gameId, nil
}

func (g *GameSer) GetRankingsSorted() ([]Ranking, error) {

	users, err := g.UserRepository.GetAll()
	rankings := make([]Ranking, 0, len(users))

	if err != nil {
		log.Printf("get all users failed %s\n", err)
		return rankings, err
	}

	users = funk.Filter(users, func(user repository.RegisteredUser) bool {
		return user.PlayedGames > 0
	}).([]repository.RegisteredUser)

	sort.SliceStable(users, func(i, j int) bool {
		if users[i].GamesWon == users[j].GamesWon {
			return users[i].BestDuration < users[j].BestDuration
		} else {
			return users[i].GamesWon > users[j].GamesWon
		}
	})

	for i, user := range users {
		i++
		rankings = append(rankings, Ranking{
			Rank:         i,
			DisplayName:  user.DisplayName,
			GamesWon:     user.GamesWon,
			BestDuration: user.BestDuration,
		})
	}

	if len(rankings) > 30 {
		return rankings[0:29], nil
	} else {
		return rankings, nil
	}
}

func (g *GameSer) RemoveGame(gameId string) (*mongo.DeleteResult, error) {

	parsedId, err := primitive.ObjectIDFromHex(gameId)

	if err != nil {
		log.Printf("can not parse game id %s\n", err)
		return nil, err
	}

	users, err := g.UserRepository.GetByGameId(parsedId)

	for _, user := range users {
		_, err := g.UserRepository.UpdateGameRelationship(&user.Id, nil)

		if err != nil {
			log.Printf("can not update game relationship %s\n", err)
			return nil, err
		}
	}

	return g.GameRepository.RemoveGame(gameId)
}

func (g *GameSer) UpdateGameState(gameId string, state repository.GameState) (*GameEntry, error) {

	currentGame, err := g.GameRepository.UpdateState(gameId, state)

	if err != nil {
		log.Printf("update game state failed %s\n", err)
		return nil, err
	}

	gameEntry := GameEntry{
		Id:           currentGame.Id.Hex(),
		Duration:     currentGame.Duration,
		KiName:       currentGame.KiName,
		KiCoins:      currentGame.KiCoins,
		Player1:      currentGame.Player1,
		Player1Coins: currentGame.Player1Coins,
		Player2:      currentGame.Player2,
		Player2Coins: currentGame.Player2Coins,
		Player3:      currentGame.Player3,
		Player3Coins: currentGame.Player3Coins,
		State:        currentGame.State,
	}

	return &gameEntry, nil
}

func (g *GameSer) UpdateCoins(player string, coins int) bool {

	if coins < 0 {
		log.Printf("coins negative. ignore\n")
		return false
	}

	currentGame, err := g.GameRepository.GetCurrent()

	if err != nil {
		log.Printf("can not find current game %s\n", err)
		return false
	}

	if strings.ToLower(player) == strings.ToLower(currentGame.Player1) {
		_, err := g.GameRepository.UpdateCoins(currentGame.Id.Hex(), repository.Player1CoinMarker, coins)
		if err != nil {
			log.Printf("update coins of player 1 failed %s\n", err)
			return false
		}
	} else if strings.ToLower(player) == strings.ToLower(currentGame.Player2) {
		_, err := g.GameRepository.UpdateCoins(currentGame.Id.Hex(), repository.Player2CoinMarker, coins)
		if err != nil {
			log.Printf("update coins of player 2 failed %s\n", err)
			return false
		}
	} else if strings.ToLower(player) == strings.ToLower(currentGame.Player3) {
		_, err := g.GameRepository.UpdateCoins(currentGame.Id.Hex(), repository.Player3CoinMarker, coins)
		if err != nil {
			log.Printf("update coins of player 3 failed %s\n", err)
			return false
		}
	} else if strings.ToLower(player) == strings.ToLower(currentGame.KiName) {
		_, err := g.GameRepository.UpdateCoins(currentGame.Id.Hex(), repository.KiCoinMarker, coins)
		if err != nil {
			log.Printf("update coins of ki failed %s\n", err)
			return false
		}
	}

	return true
}

func (g *GameSer) UpdateGameDuration(gameId string, duration float64) {

	_, err := g.GameRepository.UpdateDuration(gameId, duration)

	if err != nil {
		log.Printf("update game duration failed %s\n", err)
	}
}

func (g *GameSer) GetCurrentGame() (*GameEntry, error) {

	game, err := g.GameRepository.GetCurrent()

	if err != nil {
		log.Printf("get current game failed %s\n", err)
		return nil, err
	}

	if game == nil {
		return nil, nil
	}

	gameEntry := GameEntry{
		Id:           game.Id.Hex(),
		KiName:       game.KiName,
		KiCoins:      game.KiCoins,
		Player1:      game.Player1,
		Player1Coins: game.Player1Coins,
		Player2:      game.Player2,
		Player2Coins: game.Player2Coins,
		Player3:      game.Player3,
		Player3Coins: game.Player3Coins,
		State:        game.State,
	}

	return &gameEntry, nil
}

func (g *GameSer) GetCurrentDashboardState() (*websocket.DashboardSignal, error) {

	game, err := g.GameRepository.GetCurrent()

	if err != nil {
		log.Printf("get current game failed %s\n", err)
		return nil, err
	}

	ranking, _ := g.GetRankingsSorted()

	if game == nil || game.State == repository.GameAnnounced || game.State == repository.GameReady {
		return &websocket.DashboardSignal{
			DashboardGame:    nil,
			DashboardRanking: ToDashboardRanking(ranking),
		}, nil
	}

	gameEntry := GameEntry{
		Id:           game.Id.Hex(),
		KiName:       game.KiName,
		KiCoins:      game.KiCoins,
		Player1:      game.Player1,
		Player1Coins: game.Player1Coins,
		Player2:      game.Player2,
		Player2Coins: game.Player2Coins,
		Player3:      game.Player3,
		Player3Coins: game.Player3Coins,
		State:        game.State,
	}

	return &websocket.DashboardSignal{
		DashboardGame:    ToDashboardGameFromGameEntry(&gameEntry),
		DashboardRanking: ToDashboardRanking(ranking),
	}, nil
}

func ToDashboardGameFromGameEntry(game *GameEntry) *websocket.DashboardGame {
	return &websocket.DashboardGame{
		DocId:        game.Id,
		Duration:     int(math.Trunc(game.Duration)),
		KiName:       game.KiName,
		KiCoins:      game.KiCoins,
		Player1:      game.Player1,
		Player1Coins: game.Player1Coins,
		Player2:      game.Player2,
		Player2Coins: game.Player2Coins,
		Player3:      game.Player3,
		Player3Coins: game.Player3Coins,
		State:        string(game.State),
	}
}

func ToDashboardRanking(ranking []Ranking) []websocket.DashboardRanking {

	dashboardRanking := make([]websocket.DashboardRanking, 0, len(ranking))

	for _, rank := range ranking {
		dashboardRanking = append(dashboardRanking, websocket.DashboardRanking{
			Rank:         rank.Rank,
			DisplayName:  rank.DisplayName,
			GamesWon:     rank.GamesWon,
			BestDuration: int(math.Trunc(rank.BestDuration)),
		})
	}

	return dashboardRanking
}
