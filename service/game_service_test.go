package service

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"louie-web-administrator/repository"
	"testing"
)

func Test_GetRankings_DurationLowerThanOneSecond(t *testing.T) {

	testUserRepository := new(repository.TestUserRepository)
	gameService := GameSer{UserRepository: testUserRepository}

	testUserRepository.On("GetAll").Return([]repository.RegisteredUser{
		{
			DisplayName:  "loser",
			BestDuration: 30.50,
			GamesWon:     0,
			PlayedGames:  20,
		},
		{
			DisplayName:  "second-winner",
			BestDuration: 50.9999999,
			GamesWon:     5,
			PlayedGames:  5,
		},
		{
			DisplayName:  "winner",
			BestDuration: 10.0,
			GamesWon:     5,
			PlayedGames:  1,
		},
		{
			DisplayName:  "nobody",
			BestDuration: -1.0,
			GamesWon:     0,
			PlayedGames:  0,
		},
	}, nil)

	rankings, err := gameService.GetRankingsSorted()

	assert.NoError(t, err)
	assert.Equal(t, []Ranking{
		{
			Rank:         1,
			DisplayName:  "winner",
			GamesWon:     5,
			BestDuration: 10.0,
		},
		{
			Rank:         2,
			DisplayName:  "second-winner",
			GamesWon:     5,
			BestDuration: 50.9999999,
		},
		{
			Rank:         3,
			DisplayName:  "loser",
			GamesWon:     0,
			BestDuration: 30.50,
		},
	}, rankings)
}

func Test_GetRankings(t *testing.T) {

	testUserRepository := new(repository.TestUserRepository)
	gameService := GameSer{UserRepository: testUserRepository}

	testUserRepository.On("GetAll").Return([]repository.RegisteredUser{
		{
			DisplayName:  "loser",
			BestDuration: 50.0,
			GamesWon:     0,
			PlayedGames:  20,
		},
		{
			DisplayName:  "second-winner",
			BestDuration: 30.0,
			GamesWon:     5,
			PlayedGames:  5,
		},
		{
			DisplayName:  "winner",
			BestDuration: 20.0,
			GamesWon:     5,
			PlayedGames:  1,
		},
		{
			DisplayName:  "nobody",
			BestDuration: -1,
			GamesWon:     0,
			PlayedGames:  0,
		},
	}, nil)

	rankings, err := gameService.GetRankingsSorted()

	assert.NoError(t, err)
	assert.Equal(t, []Ranking{
		{
			Rank:         1,
			DisplayName:  "winner",
			GamesWon:     5,
			BestDuration: 20.0,
		},
		{
			Rank:         2,
			DisplayName:  "second-winner",
			GamesWon:     5,
			BestDuration: 30.0,
		},
		{
			Rank:         3,
			DisplayName:  "loser",
			GamesWon:     0,
			BestDuration: 50.0,
		},
	}, rankings)
}

func Test_GetRankings_NoUsers(t *testing.T) {

	testUserRepository := new(repository.TestUserRepository)
	gameService := GameSer{UserRepository: testUserRepository}

	testUserRepository.On("GetAll").Return([]repository.RegisteredUser{}, nil)

	rankings, err := gameService.GetRankingsSorted()

	assert.NoError(t, err)
	assert.Equal(t, []Ranking{}, rankings)
}

func Test_GetRankings_Error(t *testing.T) {

	testUserRepository := new(repository.TestUserRepository)
	gameService := GameSer{UserRepository: testUserRepository}

	testUserRepository.On("GetAll").Return([]repository.RegisteredUser{}, errors.New("new error"))

	rankings, err := gameService.GetRankingsSorted()

	assert.Error(t, errors.New("new error"), err)
	assert.Equal(t, []Ranking{}, rankings)
}
