package admin

import (
	"github.com/stretchr/testify/assert"
	"louie-web-administrator/service"
	"testing"
)

func Test_CalculatePages_MoreThanTen(t *testing.T) {

	users := make([]service.UserEntry, 0, 11)

	for i := 0; i < 11; i++ {
		users = append(users, service.UserEntry{})
	}

	testPaging := calculatePages(int64(len(users)), 1)

	assert.Equal(t, paging{Pages: []page{{Number: 1, Active: true}, {Number: 2, Active: false}}}, testPaging)
}

func Test_CalculatePages_LessThanTen(t *testing.T) {

	users := make([]service.UserEntry, 0, 7)

	for i := 0; i < 7; i++ {
		users = append(users, service.UserEntry{})
	}

	testPaging := calculatePages(int64(len(users)), 1)

	assert.Equal(t, paging{Pages: []page{{Number: 1, Active: true}}}, testPaging)
}

func Test_CalculatePages_Empty(t *testing.T) {

	users := make([]service.UserEntry, 0, 101)

	testPaging := calculatePages(int64(len(users)), 1)

	assert.Equal(t, paging{Pages: []page{}}, testPaging)
}

func Test_CalculatePages_Ten(t *testing.T) {

	users := make([]service.UserEntry, 0, 10)

	for i := 0; i < 10; i++ {
		users = append(users, service.UserEntry{})
	}

	testPaging := calculatePages(int64(len(users)), 1)

	assert.Equal(t, paging{Pages: []page{{Number: 1, Active: true}}}, testPaging)
}
