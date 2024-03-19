package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/thoas/go-funk"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"louie-web-administrator/repository"
	"slices"
	"testing"
)

func Test_ShufflePositionsForActiveUsers(t *testing.T) {

	testUserRepository := new(repository.TestUserRepository)
	userService := UserSer{UserRepository: testUserRepository}

	maxId := primitive.NewObjectID()
	janId := primitive.NewObjectID()
	homerId := primitive.NewObjectID()
	kiId := primitive.NewObjectID()

	testUserRepository.On("GetAllActive").Return([]repository.RegisteredUser{
		{
			Id:          maxId,
			DisplayName: "max",
			IsKiUser:    false,
		},
		{
			Id:          janId,
			DisplayName: "jan",
			IsKiUser:    false,
		},
		{
			Id:          homerId,
			DisplayName: "homer",
			IsKiUser:    false,
		},
		{
			Id:          kiId,
			DisplayName: "ki",
			IsKiUser:    true,
		},
	}, nil)

	testUserRepository.On("UpdatePosition", maxId.Hex(), mock.Anything).Once().Return(&mongo.UpdateResult{
		MatchedCount:  0,
		ModifiedCount: 0,
		UpsertedCount: 1,
		UpsertedID:    maxId,
	}, nil)
	testUserRepository.On("UpdatePosition", janId.Hex(), mock.Anything).Once().Return(&mongo.UpdateResult{
		MatchedCount:  0,
		ModifiedCount: 0,
		UpsertedCount: 1,
		UpsertedID:    janId,
	}, nil)
	testUserRepository.On("UpdatePosition", homerId.Hex(), mock.Anything).Once().Return(&mongo.UpdateResult{
		MatchedCount:  0,
		ModifiedCount: 0,
		UpsertedCount: 1,
		UpsertedID:    homerId,
	}, nil)

	userService.ShufflePositionsForActiveUsers()

	testUserRepository.AssertExpectations(t)

	calls := funk.Filter(testUserRepository.Calls, func(call mock.Call) bool {
		return call.Method == "UpdatePosition"
	}).([]mock.Call)

	positions := funk.Map(calls, func(call mock.Call) string {
		return call.Arguments[1].(string)
	}).([]string)

	slices.Sort(positions)

	assert.Equal(t, []string{"1", "2", "3"}, positions)
}
