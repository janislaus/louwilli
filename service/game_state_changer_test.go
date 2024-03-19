package service

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"louie-web-administrator/louie_kafka"
	"louie-web-administrator/repository"
	"louie-web-administrator/websocket"
	"math"
	"testing"
)

var (
	gameId       = primitive.NewObjectID().Hex()
	gameDuration = repository.InitialGameDuration
	kiName       = "Louie"
	kiCoins      = 3
	player1Name  = "tobi"
	player1Coins = 3
	player2Name  = "willi"
	player2Coins = 3
	player3Name  = "jann"
	player3Coins = 3

	gameState = repository.GameAnnounced
)

func Test_CheckAndUpdateGameState_SwitchToReady(t *testing.T) {

	gameService := initMockedGameService()

	testUserService := new(TestUserService)

	dashboardSocketChannel := make(chan *websocket.DashboardSignal, 100)
	adminUiChannel := make(chan websocket.AdminUiEvent, 10)

	gameDashboardSocket := websocket.GameDashboardSocket{
		GameDashboardChannel:     dashboardSocketChannel,
		GetCurrentDashboardState: gameService.GetCurrentDashboardState,
	}

	adminUiWebsocket := websocket.InitAdminUiWebsocket(adminUiChannel)

	gameStateChanger := GameStateChecker{
		UserService:         testUserService,
		GameService:         gameService,
		GameDashboardSocket: gameDashboardSocket,
		AdminUiSocket:       *adminUiWebsocket,
	}

	playersCanBeReceived, _ := json.Marshal(louie_kafka.PlayersCanBeReceivedEvent{
		Event:  louie_kafka.PlayersCanBeReceived,
		Sender: "webserver",
	})

	eventProcessed := gameStateChanger.checkAndUpdateGameState(playersCanBeReceived)

	assert.True(t, eventProcessed)

	adminEvent := <-adminUiChannel

	assert.Equal(t, websocket.AdminUiEvent{EventType: "ready", KiCoins: 3, Player1Coins: 3, Player2Coins: 3, Player3Coins: 3}, adminEvent)
}

func Test_CheckAndUpdateGameState_SwitchToReady_NoCurrentGameExists(t *testing.T) {

	testGameService := new(testGameService)
	testGameService.On("GetCurrentGame").Return(nil, nil)

	testUserService := new(TestUserService)

	dashboardSocketChannel := make(chan *websocket.DashboardSignal, 100)
	adminUiChannel := make(chan websocket.AdminUiEvent, 10)

	adminUiWebsocket := websocket.InitAdminUiWebsocket(adminUiChannel)

	gameDashboardSocket := websocket.GameDashboardSocket{
		GameDashboardChannel:     dashboardSocketChannel,
		GetCurrentDashboardState: testGameService.GetCurrentDashboardState,
	}

	gameStateChanger := GameStateChecker{
		UserService:         testUserService,
		GameService:         testGameService,
		GameDashboardSocket: gameDashboardSocket,
		AdminUiSocket:       *adminUiWebsocket,
	}

	playersCanBeReceived, _ := json.Marshal(louie_kafka.PlayersCanBeReceivedEvent{
		Event:  louie_kafka.PlayersCanBeReceived,
		Sender: "webserver",
	})

	eventProcessed := gameStateChanger.checkAndUpdateGameState(playersCanBeReceived)

	assert.False(t, eventProcessed)
}

func Test_CheckAndUpdateGameState_SwitchToReady_CurrentGameNotInAnnouncedState(t *testing.T) {

	testGameService := new(testGameService)
	testGameService.On("GetCurrentGame").Return(&GameEntry{
		Id:           gameId,
		Duration:     gameDuration,
		KiName:       kiName,
		KiCoins:      kiCoins,
		Player1:      player1Name,
		Player1Coins: player1Coins,
		Player2:      player2Name,
		Player2Coins: player2Coins,
		Player3:      player3Name,
		Player3Coins: player3Coins,
		State:        repository.GameFinished,
	}, nil)

	testUserService := new(TestUserService)

	dashboardSocketChannel := make(chan *websocket.DashboardSignal, 100)
	adminUiChannel := make(chan websocket.AdminUiEvent, 10)

	adminUiWebsocket := websocket.InitAdminUiWebsocket(adminUiChannel)

	gameDashboardSocket := websocket.GameDashboardSocket{
		GameDashboardChannel:     dashboardSocketChannel,
		GetCurrentDashboardState: testGameService.GetCurrentDashboardState,
	}

	gameStateChanger := GameStateChecker{
		UserService:         testUserService,
		GameService:         testGameService,
		GameDashboardSocket: gameDashboardSocket,
		AdminUiSocket:       *adminUiWebsocket,
	}

	playersCanBeReceived, _ := json.Marshal(louie_kafka.PlayersCanBeReceivedEvent{
		Event:  louie_kafka.PlayersCanBeReceived,
		Sender: "webserver",
	})

	eventProcessed := gameStateChanger.checkAndUpdateGameState(playersCanBeReceived)

	assert.False(t, eventProcessed)
}

func Test_CheckAndUpdateGameState_SwitchToReady_FailureDuringUpdate(t *testing.T) {

	testGameService := new(testGameService)
	testGameService.On("GetCurrentGame").Return(&GameEntry{
		Id:           gameId,
		Duration:     gameDuration,
		KiName:       kiName,
		KiCoins:      kiCoins,
		Player1:      player1Name,
		Player1Coins: player1Coins,
		Player2:      player2Name,
		Player2Coins: player2Coins,
		Player3:      player3Name,
		Player3Coins: player3Coins,
		State:        repository.GameAnnounced,
	}, nil)

	testGameService.On("UpdateGameState", gameId, repository.GameReady).Return(nil, errors.New("new error"))

	testUserService := new(TestUserService)

	dashboardSocketChannel := make(chan *websocket.DashboardSignal, 100)
	adminUiChannel := make(chan websocket.AdminUiEvent, 10)

	gameDashboardSocket := websocket.GameDashboardSocket{
		GameDashboardChannel:     dashboardSocketChannel,
		GetCurrentDashboardState: testGameService.GetCurrentDashboardState,
	}

	adminUiWebsocket := websocket.InitAdminUiWebsocket(adminUiChannel)

	gameStateChanger := GameStateChecker{
		UserService:         testUserService,
		GameService:         testGameService,
		GameDashboardSocket: gameDashboardSocket,
		AdminUiSocket:       *adminUiWebsocket,
	}

	playersCanBeReceived, _ := json.Marshal(louie_kafka.PlayersCanBeReceivedEvent{
		Event:  louie_kafka.PlayersCanBeReceived,
		Sender: "webserver",
	})

	eventProcessed := gameStateChanger.checkAndUpdateGameState(playersCanBeReceived)

	assert.False(t, eventProcessed)
}

func Test_CheckAndUpdateGameState_SwitchToActive_NoCurrentGame(t *testing.T) {

	testGameService := new(testGameService)
	testGameService.On("GetCurrentGame").Return(nil, nil)

	testUserService := new(TestUserService)

	dashboardSocketChannel := make(chan *websocket.DashboardSignal, 100)
	adminUiChannel := make(chan websocket.AdminUiEvent, 10)

	gameDashboardSocket := websocket.GameDashboardSocket{
		GameDashboardChannel:     dashboardSocketChannel,
		GetCurrentDashboardState: testGameService.GetCurrentDashboardState,
	}

	adminUiWebsocket := websocket.InitAdminUiWebsocket(adminUiChannel)

	gameStateChanger := GameStateChecker{
		UserService:         testUserService,
		GameService:         testGameService,
		GameDashboardSocket: gameDashboardSocket,
		AdminUiSocket:       *adminUiWebsocket,
	}

	playersConfirmed, _ := json.Marshal(louie_kafka.PlayersCanBeReceivedEvent{
		Event:  louie_kafka.PlayersConfirm,
		Sender: "webserver",
	})

	eventProcessed := gameStateChanger.checkAndUpdateGameState(playersConfirmed)

	assert.False(t, eventProcessed)
}

func Test_CheckAndUpdateGameState_SwitchToActive_CurrentGameNotInReadyState(t *testing.T) {

	testGameService := new(testGameService)
	testGameService.On("GetCurrentGame").Return(&GameEntry{
		Id:           gameId,
		Duration:     gameDuration,
		KiName:       kiName,
		KiCoins:      kiCoins,
		Player1:      player1Name,
		Player1Coins: player1Coins,
		Player2:      player2Name,
		Player2Coins: player2Coins,
		Player3:      player3Name,
		Player3Coins: player3Coins,
		State:        repository.GameFinished,
	}, nil)

	testUserService := new(TestUserService)

	dashboardSocketChannel := make(chan *websocket.DashboardSignal, 100)
	adminUiChannel := make(chan websocket.AdminUiEvent, 10)

	gameDashboardSocket := websocket.GameDashboardSocket{
		GameDashboardChannel:     dashboardSocketChannel,
		GetCurrentDashboardState: testGameService.GetCurrentDashboardState,
	}

	adminUiWebsocket := websocket.InitAdminUiWebsocket(adminUiChannel)

	gameStateChanger := GameStateChecker{
		UserService:         testUserService,
		GameService:         testGameService,
		GameDashboardSocket: gameDashboardSocket,
		AdminUiSocket:       *adminUiWebsocket,
	}

	playersConfirmed, _ := json.Marshal(louie_kafka.PlayersCanBeReceivedEvent{
		Event:  louie_kafka.PlayersConfirm,
		Sender: "webserver",
	})

	eventProcessed := gameStateChanger.checkAndUpdateGameState(playersConfirmed)

	assert.False(t, eventProcessed)
}

func Test_CheckAndUpdateGameState_SwitchToActive(t *testing.T) {

	testGameService := new(testGameService)
	testGameService.On("GetCurrentGame").Return(&GameEntry{
		Id:           gameId,
		Duration:     gameDuration,
		KiName:       kiName,
		KiCoins:      kiCoins,
		Player1:      player1Name,
		Player1Coins: player1Coins,
		Player2:      player2Name,
		Player2Coins: player2Coins,
		Player3:      player3Name,
		Player3Coins: player3Coins,
		State:        repository.GameReady,
	}, nil)

	testGameService.On("GetRankingsSorted").Return(
		[]Ranking{
			{
				Rank:         1,
				DisplayName:  "willi",
				GamesWon:     5000,
				BestDuration: 5000,
			},
		}, nil)

	testGameService.On("UpdateGameState", gameId, repository.GameActive).Return(&GameEntry{
		Id:           gameId,
		Duration:     gameDuration,
		KiName:       kiName,
		KiCoins:      kiCoins,
		Player1:      player1Name,
		Player1Coins: player1Coins,
		Player2:      player2Name,
		Player2Coins: player2Coins,
		Player3:      player3Name,
		Player3Coins: player3Coins,
		State:        repository.GameActive,
	}, nil)

	testUserService := new(TestUserService)

	dashboardSocketChannel := make(chan *websocket.DashboardSignal, 100)
	adminUiChannel := make(chan websocket.AdminUiEvent, 10)

	gameDashboardSocket := websocket.GameDashboardSocket{
		GameDashboardChannel:     dashboardSocketChannel,
		GetCurrentDashboardState: testGameService.GetCurrentDashboardState,
	}

	adminUiWebsocket := websocket.InitAdminUiWebsocket(adminUiChannel)

	gameStateChanger := GameStateChecker{
		UserService:         testUserService,
		GameService:         testGameService,
		GameDashboardSocket: gameDashboardSocket,
		AdminUiSocket:       *adminUiWebsocket,
	}

	playersConfirmed, _ := json.Marshal(louie_kafka.PlayersCanBeReceivedEvent{
		Event:  louie_kafka.PlayersConfirm,
		Sender: "webserver",
	})

	eventProcessed := gameStateChanger.checkAndUpdateGameState(playersConfirmed)
	adminSignal := <-adminUiChannel
	dashboardSignal := <-dashboardSocketChannel

	assert.True(t, eventProcessed)
	assert.Equal(t, websocket.AdminUiEvent{EventType: "active", KiCoins: 3, Player1Coins: 3, Player2Coins: 3, Player3Coins: 3}, adminSignal)
	assert.Equal(t, &websocket.DashboardSignal{
		DashboardGame: &websocket.DashboardGame{
			DocId:        gameId,
			Duration:     int(math.Trunc(gameDuration)),
			KiName:       kiName,
			KiCoins:      kiCoins,
			Player1:      player1Name,
			Player1Coins: player1Coins,
			Player2:      player2Name,
			Player2Coins: player2Coins,
			Player3:      player3Name,
			Player3Coins: player3Coins,
			State:        "active",
		},
		DashboardRanking: []websocket.DashboardRanking{{Rank: 1, DisplayName: "willi", GamesWon: 5000, BestDuration: 5000}},
	}, dashboardSignal)
}
func Test_CheckAndUpdateGameState_SwitchToFinished_NoCurrentGame(t *testing.T) {

	testGameService := new(testGameService)
	testGameService.On("GetCurrentGame").Return(nil, nil)

	testUserService := new(TestUserService)

	dashboardSocketChannel := make(chan *websocket.DashboardSignal, 100)
	adminUiChannel := make(chan websocket.AdminUiEvent, 10)

	gameDashboardSocket := websocket.GameDashboardSocket{
		GameDashboardChannel:     dashboardSocketChannel,
		GetCurrentDashboardState: testGameService.GetCurrentDashboardState,
	}

	adminUiWebsocket := websocket.InitAdminUiWebsocket(adminUiChannel)

	gameStateChanger := GameStateChecker{
		UserService:         testUserService,
		GameService:         testGameService,
		GameDashboardSocket: gameDashboardSocket,
		AdminUiSocket:       *adminUiWebsocket,
	}

	gameDone, _ := json.Marshal(louie_kafka.PlayersCanBeReceivedEvent{
		Event:  louie_kafka.GameDone,
		Sender: "webserver",
	})

	eventProcessed := gameStateChanger.checkAndUpdateGameState(gameDone)

	assert.False(t, eventProcessed)
}

func Test_CheckAndUpdateGameState_SwitchToFinished_CurrentGameNotInActiveState(t *testing.T) {

	testGameService := new(testGameService)
	testGameService.On("GetCurrentGame").Return(&GameEntry{
		Id:           gameId,
		Duration:     gameDuration,
		KiName:       kiName,
		KiCoins:      kiCoins,
		Player1:      player1Name,
		Player1Coins: player1Coins,
		Player2:      player2Name,
		Player2Coins: player2Coins,
		Player3:      player3Name,
		Player3Coins: player3Coins,
		State:        repository.GameFinished,
	}, nil)

	testUserService := new(TestUserService)

	dashboardSocketChannel := make(chan *websocket.DashboardSignal, 100)
	adminUiChannel := make(chan websocket.AdminUiEvent, 10)

	gameDashboardSocket := websocket.GameDashboardSocket{
		GameDashboardChannel:     dashboardSocketChannel,
		GetCurrentDashboardState: testGameService.GetCurrentDashboardState,
	}

	adminUiWebsocket := websocket.InitAdminUiWebsocket(adminUiChannel)

	gameStateChanger := GameStateChecker{
		UserService:         testUserService,
		GameService:         testGameService,
		GameDashboardSocket: gameDashboardSocket,
		AdminUiSocket:       *adminUiWebsocket,
	}

	gameDone, _ := json.Marshal(louie_kafka.PlayersCanBeReceivedEvent{
		Event:  louie_kafka.GameDone,
		Sender: "webserver",
	})

	eventProcessed := gameStateChanger.checkAndUpdateGameState(gameDone)

	assert.False(t, eventProcessed)
}

func initMockedGameService() *testGameService {
	testGameService := new(testGameService)

	testGameService.On("GetCurrentDashboardState").Return(
		websocket.DashboardSignal{
			DashboardGame: ToDashboardGameFromGameEntry(&GameEntry{
				Id:           gameId,
				Duration:     gameDuration,
				KiName:       kiName,
				KiCoins:      kiCoins,
				Player1:      player1Name,
				Player1Coins: player1Coins,
				Player2:      player2Name,
				Player2Coins: player2Coins,
				Player3:      player3Name,
				Player3Coins: player3Coins,
				State:        gameState,
			}),
			DashboardRanking: nil,
		}, nil)

	testGameService.On("SendPlayerReadyMessageToKafka", []louie_kafka.PlayerDisplayName{
		{"tobi"}, {"willi"}, {"jann"},
	}).Return()

	testGameService.On("UpdateGameState", gameId, repository.GameReady).Return(&GameEntry{
		Id:           gameId,
		Duration:     gameDuration,
		KiName:       kiName,
		KiCoins:      kiCoins,
		Player1:      player1Name,
		Player1Coins: player1Coins,
		Player2:      player2Name,
		Player2Coins: player2Coins,
		Player3:      player3Name,
		Player3Coins: player3Coins,
		State:        repository.GameReady,
	}, nil)

	testGameService.On("GetCurrentGame").Return(&GameEntry{
		Id:           gameId,
		Duration:     gameDuration,
		KiName:       kiName,
		KiCoins:      kiCoins,
		Player1:      player1Name,
		Player1Coins: player1Coins,
		Player2:      player2Name,
		Player2Coins: player2Coins,
		Player3:      player3Name,
		Player3Coins: player3Coins,
		State:        gameState,
	}, nil)

	return testGameService
}
