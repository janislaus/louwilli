package service

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"louie-web-administrator/repository"
	"testing"
	"time"
)

func Test_FilterNewActiveUsers_ExistingActiveUser(t *testing.T) {

	userId := primitive.NewObjectID()

	newUsersForStateUpdate := filterNewActiveUsers([]Tuple{
		{Key: userId.Hex(), Value: "active"},
	}, []repository.RegisteredUser{
		{
			Id: userId,
		},
	})

	assert.Equal(t, []string{}, newUsersForStateUpdate)
}

func Test_FilterNewActiveUsers_NewActiveUser(t *testing.T) {

	userId := primitive.NewObjectID()

	newUsersForStateUpdate := filterNewActiveUsers([]Tuple{
		{Key: userId.Hex(), Value: "active"},
	}, []repository.RegisteredUser{})

	assert.Equal(t, []string{userId.Hex()}, newUsersForStateUpdate)
}

func Test_FilterNewActiveUsers_NoNewActive(t *testing.T) {

	userId := primitive.NewObjectID()

	newUsersForStateUpdate := filterNewActiveUsers([]Tuple{},
		[]repository.RegisteredUser{
			{Id: userId},
		})

	assert.Equal(t, []string{}, newUsersForStateUpdate)
}
func Test_FilterTuplesByValue(t *testing.T) {

	filteredTuples := filterTuplesByValue([]Tuple{
		{Key: "1", Value: "active"},
		{Key: "2", Value: "waiting"},
		{Key: "3", Value: "waiting"},
	}, "active")

	assert.Equal(t, []Tuple{
		{Key: "1", Value: "active"},
	}, filteredTuples)
}
func Test_MapTuplesToIds_EmptyTuple(t *testing.T) {

	ids := mapTuplesToIds([]Tuple{})

	assert.Equal(t, []string{}, ids)
}
func Test_MapTuplesToIds(t *testing.T) {

	ids := mapTuplesToIds([]Tuple{
		{Key: "1", Value: "2"},
		{Key: "2", Value: "22"},
		{Key: "3", Value: "222"},
	})

	assert.Equal(t, []string{"1", "2", "3"}, ids)
}

func Test_ToTuples_EmptyMap(t *testing.T) {

	emptyMap := make(map[string][]string)

	tuples, err := mapToTuples(emptyMap, "", "")

	assert.NoError(t, err)
	assert.Empty(t, tuples)
}

func Test_ToTuples_MapWithTwoEntries(t *testing.T) {

	tuples, err := mapToTuples(map[string][]string{
		"id":       {"1", "11", "111"},
		"position": {"2", "22", "222"},
	}, "id", "position")

	assert.NoError(t, err)
	assert.Equal(t, []Tuple{
		{Key: "1", Value: "2"},
		{Key: "11", Value: "22"},
		{Key: "111", Value: "222"},
	}, tuples)
}

func Test_ToTuples_MapWithTwoEntry_WrongIndex(t *testing.T) {

	tuples, err := mapToTuples(map[string][]string{
		"id":       {"1", "11", "111"},
		"position": {"2", "22", "222"},
	}, "bla", "foo")

	assert.NoError(t, err)
	assert.Equal(t, []Tuple{}, tuples)
}
func Test_MapTuples(t *testing.T) {

	formParameter := map[string][]string{
		"id": {
			"#1",
		},
		"position": {
			"1",
		},
	}

	tuples, err := mapToTuples(formParameter, "id", "position")

	assert.NoError(t, err)
	assert.Equal(t, []Tuple{
		{Key: "#1", Value: "1"},
	}, tuples)

}
func Test_SortByDate(t *testing.T) {

	timeBefore := time.Date(2023, 11, 24, 0, 0, 0, 0, time.UTC)
	timeAfter := time.Date(2023, 11, 24, 1, 0, 0, 0, time.UTC)

	users := []repository.RegisteredUser{
		{
			RegistrationTimestamp: timeBefore,
		},
		{
			RegistrationTimestamp: timeAfter,
		},
	}

	sortByDateDescending(users)

	assert.Equal(t, []repository.RegisteredUser{
		{
			RegistrationTimestamp: timeAfter,
		},
		{
			RegistrationTimestamp: timeBefore,
		},
	}, users)
}

func Test_MapUsersToIds(t *testing.T) {

	userId := primitive.NewObjectID()

	users := []repository.RegisteredUser{
		{
			Id: userId,
		},
	}

	ids := mapUsersToIds(users)

	assert.Equal(t, []string{userId.Hex()}, ids)
}

func Test_MapUsersToIds_EmptySlice(t *testing.T) {

	users := []repository.RegisteredUser{}

	ids := mapUsersToIds(users)

	assert.Equal(t, []string{}, ids)
}

func Test_MapUsersToIds_Nil(t *testing.T) {

	var users []repository.RegisteredUser

	ids := mapUsersToIds(users)

	assert.Equal(t, []string{}, ids)
}
