package repository

import (
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TestUserRepository struct {
	mock.Mock
}

func (testUserRepository *TestUserRepository) UpdateGameStatisticValues(user RegisteredUser) (*mongo.UpdateResult, error) {
	args := testUserRepository.Called(user)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (testUserRepository *TestUserRepository) UpdateGameRelationship(userId *primitive.ObjectID, gameId *primitive.ObjectID) (*mongo.UpdateResult, error) {
	args := testUserRepository.Called(userId, gameId)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (testUserRepository *TestUserRepository) GetByGameId(gameId primitive.ObjectID) ([]RegisteredUser, error) {
	args := testUserRepository.Called(gameId)
	return args.Get(0).([]RegisteredUser), args.Error(1)
}

func (testUserRepository *TestUserRepository) Create(user RegisteredUser) (*mongo.InsertOneResult, error) {
	args := testUserRepository.Called(user)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (testUserRepository *TestUserRepository) Get(email string) (*RegisteredUser, error) {
	args := testUserRepository.Called(email)
	return args.Get(0).(*RegisteredUser), args.Error(1)
}

func (testUserRepository *TestUserRepository) UpdatePosition(id string, position string) (*mongo.UpdateResult, error) {
	args := testUserRepository.Called(id, position)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (testUserRepository *TestUserRepository) UpdateState(id string, state string) (*mongo.UpdateResult, error) {
	args := testUserRepository.Called(id, state)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (testUserRepository *TestUserRepository) GetAllActive() ([]RegisteredUser, error) {
	args := testUserRepository.Called()
	return args.Get(0).([]RegisteredUser), args.Error(1)
}

func (testUserRepository *TestUserRepository) GetAll() ([]RegisteredUser, error) {
	args := testUserRepository.Called()
	return args.Get(0).([]RegisteredUser), args.Error(1)
}

func (testUserRepository *TestUserRepository) GetPagedSortedByRegistrationDateWithoutKiUser(page int64, nameFilter string) ([]RegisteredUser, error) {
	args := testUserRepository.Called(page, nameFilter)
	return args.Get(0).([]RegisteredUser), args.Error(1)
}

func (testUserRepository *TestUserRepository) Remove(displayName string) error {
	args := testUserRepository.Called(displayName)
	return args.Error(1)
}
func (testUserRepository *TestUserRepository) GetByDisplayName(displayName string) (*RegisteredUser, error) {
	args := testUserRepository.Called(displayName)
	return args.Get(0).(*RegisteredUser), args.Error(1)
}

func (testUserRepository *TestUserRepository) CreateKiUser() (*mongo.InsertOneResult, error) {
	args := testUserRepository.Called()
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (testUserRepository *TestUserRepository) CountAllWithoutKiUser(nameFilter string) (int64, error) {
	args := testUserRepository.Called(nameFilter)
	return args.Get(0).(int64), args.Error(1)
}

func (testUserRepository *TestUserRepository) UpdateAllNonKiUsers(state UserState) (*mongo.UpdateResult, error) {
	args := testUserRepository.Called(state)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}
