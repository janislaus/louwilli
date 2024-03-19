package service

import (
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"louie-web-administrator/repository"
)

type TestUserService struct {
	mock.Mock
}

func (testUserService *TestUserService) Create(dashboardUser *DashboardUser) (*mongo.InsertOneResult, error) {
	args := testUserService.Called(dashboardUser)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (testUserService *TestUserService) GetAllActive() ([]repository.RegisteredUser, error) {
	args := testUserService.Called()
	return args.Get(0).([]repository.RegisteredUser), args.Error(1)
}

func (testUserService *TestUserService) GetUsers(page int64, nameFilter string) []UserEntry {
	args := testUserService.Called(page, nameFilter)
	return args.Get(0).([]UserEntry)
}

func (testUserService *TestUserService) CountActiveUsersWithoutKiUser() int {
	args := testUserService.Called()
	return args.Get(0).(int)
}

func (testUserService *TestUserService) CountAllWithoutKiUser(nameFilter string) int64 {
	args := testUserService.Called(nameFilter)
	return args.Get(0).(int64)
}

func (testUserService *TestUserService) UpdateStatistic(user repository.RegisteredUser) (*mongo.UpdateResult, error) {
	args := testUserService.Called(user)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (testUserService *TestUserService) UpdatePosition(id string, position string) (int64, error) {
	args := testUserService.Called(id, position)
	return args.Get(0).(int64), args.Error(1)
}

func (testUserService *TestUserService) UpdateState(id string, state string) (int64, error) {
	args := testUserService.Called(id, state)
	return args.Get(0).(int64), args.Error(1)
}

func (testUserService *TestUserService) GetByGameId(gameId primitive.ObjectID) ([]repository.RegisteredUser, error) {
	args := testUserService.Called(gameId)
	return args.Get(0).([]repository.RegisteredUser), args.Error(1)
}

func (testUserService *TestUserService) MapUserIdsToPositions(formParameters map[string][]string) ([]Tuple, error) {
	args := testUserService.Called(formParameters)
	return args.Get(0).([]Tuple), args.Error(1)
}

func (testUserService *TestUserService) MapUserIdsToStates(formParameters map[string][]string) ([]Tuple, error) {
	args := testUserService.Called(formParameters)
	return args.Get(0).([]Tuple), args.Error(1)
}

func (testUserService *TestUserService) FilterNewUserIdsForActivation(userIdsToStates []Tuple) []string {
	args := testUserService.Called(userIdsToStates)
	return args.Get(0).([]string)
}

func (testUserService *TestUserService) InitOrRefreshLouki() {
	testUserService.Called()
	return
}

func (testUserService *TestUserService) SetAllToWaiting() error {
	args := testUserService.Called()
	return args.Error(0)
}

func (testUserService *TestUserService) ShufflePositionsForActiveUsers() {
	testUserService.Called()
	return
}
