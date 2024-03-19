package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
)

func (s *RepositoryTestSuite) Test_CreateGame() {

	gameRepository := NewGameRepository(context.Background(), s.mongoDb.Client(), s.mongoDb.Name())

	gameId, err := gameRepository.CreateGame([]RegisteredUser{
		{
			DisplayName: "max",
			Pos:         "1",
		},
		{
			DisplayName: "emil",
			Pos:         "2",
		},
		{
			DisplayName: "andreas",
			Pos:         "3",
		},
	})

	assert.NoError(s.T(), err)

	currentGame, err := gameRepository.GetCurrent()
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), &GameEntity{
		Id:           *gameId,
		Duration:     InitialGameDuration,
		KiName:       KiName,
		KiCoins:      3,
		Player1:      "max",
		Player1Coins: 3,
		Player2:      "emil",
		Player2Coins: 3,
		Player3:      "andreas",
		Player3Coins: 3,
		State:        GameAnnounced,
	}, currentGame)
}

func (s *RepositoryTestSuite) Test_CreateGame_WithoutUsers() {

	gameRepository := NewGameRepository(context.Background(), s.mongoDb.Client(), s.mongoDb.Name())

	gameId, err := gameRepository.CreateGame([]RegisteredUser{})

	assert.NoError(s.T(), err)

	currentGame, err := gameRepository.GetCurrent()
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), &GameEntity{
		Id:           *gameId,
		Duration:     InitialGameDuration,
		KiName:       KiName,
		KiCoins:      3,
		Player1:      "",
		Player1Coins: 0,
		Player2:      "",
		Player2Coins: 0,
		Player3:      "",
		Player3Coins: 0,
		State:        GameAnnounced,
	}, currentGame)
}
