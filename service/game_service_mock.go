package service

import (
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"louie-web-administrator/louie_kafka"
	"louie-web-administrator/repository"
	"louie-web-administrator/websocket"
)

type testGameService struct {
	mock.Mock
}

func (testGameService *testGameService) SendPlayerReadyMessageToKafka(playerDisplayNames []louie_kafka.PlayerDisplayName) {
	testGameService.Called(playerDisplayNames)
	return
}

func (testGameService *testGameService) CreateGame(gameMembers []repository.RegisteredUser) (*primitive.ObjectID, error) {
	args := testGameService.Called(gameMembers)
	return args.Get(0).(*primitive.ObjectID), args.Error(1)
}

func (testGameService *testGameService) RemoveGame(gameId string) (*mongo.DeleteResult, error) {
	args := testGameService.Called(gameId)
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (testGameService *testGameService) UpdateGameState(gameId string, state repository.GameState) (*GameEntry, error) {
	args := testGameService.Called(gameId, state)

	get := args.Get(0)

	if get != nil {
		return get.(*GameEntry), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func (testGameService *testGameService) GetCurrentGame() (*GameEntry, error) {
	args := testGameService.Called()

	get := args.Get(0)

	if get != nil {
		return get.(*GameEntry), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func (testGameService *testGameService) UpdateGameDuration(gameId string, duration float64) {
	testGameService.Called(gameId, duration)
	return
}
func (testGameService *testGameService) UpdateCoins(player string, coins int) bool {
	args := testGameService.Called(player, coins)
	return args.Get(0).(bool)
}

func (testGameService *testGameService) GetCurrentDashboardState() (*websocket.DashboardSignal, error) {
	args := testGameService.Called()
	return args.Get(0).(*websocket.DashboardSignal), args.Error(1)
}

func (testGameService *testGameService) GetRankingsSorted() ([]Ranking, error) {
	args := testGameService.Called()
	return args.Get(0).([]Ranking), args.Error(1)
}
