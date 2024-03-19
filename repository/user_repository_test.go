package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func (s *RepositoryTestSuite) Test_Remove_NonExistingUser() {

	userRepository := NewUserRepo(context.Background(), s.mongoDb.Client(), s.mongoDb.Name())
	err := userRepository.Remove("foobar")

	assert.NoError(s.T(), err)
}

func (s *RepositoryTestSuite) Test_Remove() {

	userRepository := NewUserRepo(context.Background(), s.mongoDb.Client(), s.mongoDb.Name())

	userRepository.Create(RegisteredUser{
		AcceptNewsletter:   true,
		AcceptNotification: true,
		DisplayName:        "max",
		Email:              "max@mustermann.de",
		FirstName:          "max",
		LastName:           "müller",
		BestDuration:       0,
		GamesWon:           0,
		PlayedGames:        0,
		Pos:                "1",
		IsKiUser:           false,
	})

	err := userRepository.Remove("max")

	assert.NoError(s.T(), err)

	registeredUser, err := userRepository.Get("max@mustermann.de")

	assert.Error(s.T(), err)
	assert.Nil(s.T(), registeredUser)
}

func (s *RepositoryTestSuite) Test_CreateKiUser() {

	userRepository := NewUserRepo(context.Background(), s.mongoDb.Client(), s.mongoDb.Name())

	_, err := userRepository.CreateKiUser()

	assert.NoError(s.T(), err)

	kiUser, err := userRepository.GetByDisplayName(KiName)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), false, kiUser.AcceptNewsletter)
	assert.Equal(s.T(), false, kiUser.AcceptNotification)
	assert.Equal(s.T(), KiName, kiUser.DisplayName)
	assert.Equal(s.T(), "", kiUser.Email)
	assert.Equal(s.T(), "", kiUser.FirstName)
	assert.Equal(s.T(), "", kiUser.LastName)
	assert.Equal(s.T(), InitialUserDuration, kiUser.BestDuration)
	assert.Equal(s.T(), 0, kiUser.GamesWon)
	assert.Equal(s.T(), 0, kiUser.PlayedGames)
	assert.Equal(s.T(), "-1", kiUser.Pos)
	assert.Equal(s.T(), true, kiUser.IsKiUser)
}

func (s *RepositoryTestSuite) Test_UpdatePosition() {

	userRepository := NewUserRepo(context.Background(), s.mongoDb.Client(), s.mongoDb.Name())

	createUserResult, _ := userRepository.Create(RegisteredUser{
		AcceptNewsletter:   true,
		AcceptNotification: true,
		DisplayName:        "max",
		Email:              "max@mustermann.de",
		FirstName:          "max",
		LastName:           "müller",
		BestDuration:       0,
		GamesWon:           0,
		PlayedGames:        0,
		Pos:                "1",
		IsKiUser:           false,
	})

	userId := createUserResult.InsertedID.(primitive.ObjectID)
	_, err := userRepository.UpdatePosition(userId.Hex(), "2")

	assert.NoError(s.T(), err)

	users, err := userRepository.GetPagedSortedByRegistrationDateWithoutKiUser(1, "max")

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "2", users[0].Pos)
}

func (s *RepositoryTestSuite) Test_UpdatePosition_UseWrongId() {

	userRepository := NewUserRepo(context.Background(), s.mongoDb.Client(), s.mongoDb.Name())

	_, err := userRepository.UpdatePosition("foobar", "2")

	assert.Error(s.T(), err)
}

func (s *RepositoryTestSuite) Test_GetPagedSortedByRegistrationDateWithoutKiUser_EmptyFilter() {

	userRepository := NewUserRepo(context.Background(), s.mongoDb.Client(), s.mongoDb.Name())

	registeredUsers, err := userRepository.GetPagedSortedByRegistrationDateWithoutKiUser(1, "")

	assert.NoError(s.T(), err)
	assert.Empty(s.T(), registeredUsers)
}
func (s *RepositoryTestSuite) Test_GetPagedSortedByRegistrationDateWithoutKiUser() {

	userRepository := NewUserRepo(context.Background(), s.mongoDb.Client(), s.mongoDb.Name())

	userRepository.Create(RegisteredUser{
		AcceptNewsletter:   true,
		AcceptNotification: true,
		DisplayName:        "max",
		Email:              "max@mustermann.de",
		FirstName:          "max",
		LastName:           "müller",
		BestDuration:       0,
		GamesWon:           0,
		PlayedGames:        0,
		Pos:                "1",
		IsKiUser:           false,
	})

	userRepository.Create(RegisteredUser{
		AcceptNewsletter:   true,
		AcceptNotification: true,
		DisplayName:        "max",
		Email:              "max@musterfrau.de",
		FirstName:          "max",
		LastName:           "müller",
		BestDuration:       0,
		GamesWon:           0,
		PlayedGames:        0,
		Pos:                "1",
		IsKiUser:           false,
	})

	registeredUsers, err := userRepository.GetPagedSortedByRegistrationDateWithoutKiUser(1, "mustermann")

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "max@mustermann.de", registeredUsers[0].Email)
}

func (s *RepositoryTestSuite) Test_GetPagedSortedByRegistrationDateWithoutKiUser_EnsureNoKiUserInResult() {

	userRepository := NewUserRepo(context.Background(), s.mongoDb.Client(), s.mongoDb.Name())

	newUser := RegisteredUser{
		Id:                    primitive.NewObjectID(),
		RegistrationTimestamp: time.Now().UTC(),

		AcceptNewsletter:   true,
		AcceptNotification: true,
		DisplayName:        "max",
		Email:              "max@mustermann.de",
		FirstName:          "max",
		LastName:           "müller",

		BestDuration: InitialUserDuration,
		GamesWon:     0,
		PlayedGames:  0,
		State:        UserWaiting,
		Pos:          "1",

		IsKiUser: true,
	}

	userRepository.collection.InsertOne(context.Background(), &newUser)

	registeredUsers, err := userRepository.GetPagedSortedByRegistrationDateWithoutKiUser(1, "mustermann")

	assert.NoError(s.T(), err)
	assert.Empty(s.T(), registeredUsers)
}

func (s *RepositoryTestSuite) TestNewUser() {

	userRepository := NewUserRepo(context.Background(), s.mongoDb.Client(), s.mongoDb.Name())

	userEmail := "max@gmail.com"

	_, err := userRepository.Create(RegisteredUser{
		AcceptNewsletter:   true,
		AcceptNotification: true,
		DisplayName:        "max",
		Email:              userEmail,
		FirstName:          "max",
		LastName:           "müller",
		BestDuration:       0,
		GamesWon:           0,
		PlayedGames:        0,
		Pos:                "1",
	})

	assert.NoError(s.T(), err)

	get, err := userRepository.Get(userEmail)

	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), get)
	assert.Equal(s.T(), true, get.AcceptNotification)
	assert.Equal(s.T(), true, get.AcceptNotification)
	assert.Equal(s.T(), "max", get.DisplayName)
	assert.Equal(s.T(), userEmail, get.Email)
	assert.Equal(s.T(), "max", get.FirstName)
	assert.Equal(s.T(), "müller", get.LastName)
	assert.Equal(s.T(), -1.0, get.BestDuration)
	assert.Equal(s.T(), 0, get.GamesWon)
	assert.Equal(s.T(), 0, get.PlayedGames)
	assert.Equal(s.T(), UserWaiting, get.State)
	assert.Equal(s.T(), "1", get.Pos)
}

func (s *RepositoryTestSuite) TestNewUserWithConflict() {

	userRepository := NewUserRepo(context.Background(), s.mongoDb.Client(), s.mongoDb.Name())

	userEmail := "max@gmail.com"

	_, err := userRepository.Create(RegisteredUser{
		AcceptNewsletter:   true,
		AcceptNotification: true,
		DisplayName:        "max",
		Email:              userEmail,
		FirstName:          "max",
		LastName:           "müller",
		BestDuration:       0,
		GamesWon:           0,
		PlayedGames:        0,
		Pos:                "0",
	})

	assert.NoError(s.T(), err)

	_, err = userRepository.Create(RegisteredUser{
		AcceptNewsletter:   true,
		AcceptNotification: true,
		DisplayName:        "max",
		Email:              userEmail,
		FirstName:          "max",
		LastName:           "müller",
		BestDuration:       0,
		GamesWon:           0,
		PlayedGames:        0,
		Pos:                "0",
	})

	assert.Error(s.T(), err)
	assert.Containsf(s.T(), err.Error(), "E11000 duplicate key error collection: test-db.registeredUsers index: email_1 dup key", "")
}
