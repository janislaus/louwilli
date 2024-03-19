package service

import (
	"fmt"
	"github.com/thoas/go-funk"
	"louie-web-administrator/repository"
	"sort"
)

type Tuple struct {
	Key   string
	Value string
}

func mapUsersToIds(users []repository.RegisteredUser) []string {
	return funk.Map(users, func(user repository.RegisteredUser) string { return user.Id.Hex() }).([]string)
}
func mapToTuples(idToPositionUrlValues map[string][]string, firstIndex string, secondIndex string) ([]Tuple, error) {
	tuples := make([]Tuple, 0)

	tupleList := funk.Zip(idToPositionUrlValues[firstIndex], idToPositionUrlValues[secondIndex])

	for _, tuple := range tupleList {
		tuples = append(tuples, Tuple{
			Key:   fmt.Sprintf("%v", tuple.Element1),
			Value: fmt.Sprintf("%v", tuple.Element2),
		})
	}

	return tuples, nil
}
func mapTuplesToIds(userIdsToStates []Tuple) []string {
	return funk.Map(userIdsToStates, func(userIdToState Tuple) string {
		return userIdToState.Key
	}).([]string)
}

func filterNonKiUsers(registeredUsers []repository.RegisteredUser) []repository.RegisteredUser {
	return funk.Filter(registeredUsers, func(user repository.RegisteredUser) bool {
		return user.IsKiUser == false
	}).([]repository.RegisteredUser)
}

func filterNewActiveUsers(userIdsToStates []Tuple, existingActiveUsers []repository.RegisteredUser) []string {
	activeUserIds := mapUsersToIds(existingActiveUsers)
	userIdsForActivation := mapTuplesToIds(filterTuplesByValue(userIdsToStates, "active"))

	return funk.Filter(userIdsForActivation, func(userId string) bool {
		return !funk.Contains(activeUserIds, userId)
	}).([]string)
}
func filterTuplesByValue(userIdsToStates []Tuple, filterValue string) []Tuple {
	return funk.Filter(userIdsToStates, func(userIdToState Tuple) bool {
		return userIdToState.Value == filterValue
	}).([]Tuple)
}
func sortByDateDescending(registeredUsers []repository.RegisteredUser) {
	sort.Slice(registeredUsers, func(i, j int) bool {
		return registeredUsers[i].RegistrationTimestamp.After(registeredUsers[j].RegistrationTimestamp)
	})
}
