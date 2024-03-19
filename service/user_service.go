package service

import (
	"github.com/thoas/go-funk"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"louie-web-administrator/repository"
)

type UserService interface {
	InitOrRefreshLouki()
	Create(dashboardUser *DashboardUser) (*mongo.InsertOneResult, error)
	GetAllActive() ([]repository.RegisteredUser, error)
	GetUsers(page int64, nameFilter string) []UserEntry
	GetByGameId(gameId primitive.ObjectID) ([]repository.RegisteredUser, error)
	CountActiveUsersWithoutKiUser() int
	CountAllWithoutKiUser(nameFilter string) int64
	ShufflePositionsForActiveUsers()
	UpdateStatistic(user repository.RegisteredUser) (*mongo.UpdateResult, error)
	UpdateState(id string, state string) (int64, error)
	UpdatePosition(id string, position string) (int64, error)
	SetAllToWaiting() error
	MapUserIdsToPositions(formParameters map[string][]string) ([]Tuple, error)
	MapUserIdsToStates(formParameters map[string][]string) ([]Tuple, error)
	FilterNewUserIdsForActivation(userIdsToStates []Tuple) []string
}

type Ranking struct {
	Rank         int
	DisplayName  string
	GamesWon     int
	BestDuration float64
}

type UserEntry struct {
	Id                  string
	RegisteredTimestamp string
	LastTimePlayed      *string
	DisplayName         string
	Email               string
	PlayedGames         int
	Pos                 string
	State               repository.UserState
}

type DashboardUser struct {
	AcceptNewsletter   bool
	AcceptNotification bool
	DisplayName        string
	Email              string
	FirstName          string
	LastName           string
}

type UserSer struct {
	UserRepository repository.UserRepository
}

func (u *UserSer) InitOrRefreshLouki() {

	kiUser, _ := u.UserRepository.GetByDisplayName(repository.KiName)

	if kiUser != nil {
		u.UserRepository.Remove(repository.KiName)
	}

	u.UserRepository.CreateKiUser()
}

func (u *UserSer) Create(dashboardUser *DashboardUser) (*mongo.InsertOneResult, error) {
	return u.UserRepository.Create(*dashboardUser.toUserEntity())
}
func (u *UserSer) GetAllActive() ([]repository.RegisteredUser, error) {
	return u.UserRepository.GetAllActive()
}

func (u *UserSer) SetAllToWaiting() error {

	_, err := u.UserRepository.UpdateAllNonKiUsers(repository.UserWaiting)

	return err
}

func (u *UserSer) GetUsers(page int64, nameFilter string) []UserEntry {
	userTemplateRows := make([]UserEntry, 0)
	paged, _ := u.UserRepository.GetPagedSortedByRegistrationDateWithoutKiUser(page, nameFilter)

	for _, registeredUser := range paged {
		userTemplateRows = append(userTemplateRows, *u.toUserEntry(&registeredUser))
	}

	return userTemplateRows
}
func (u *UserSer) CountAllWithoutKiUser(nameFilter string) int64 {

	usersCount, _ := u.UserRepository.CountAllWithoutKiUser(nameFilter)

	return usersCount
}
func (u *UserSer) CountActiveUsersWithoutKiUser() int {
	activeUsers, _ := u.GetAllActive()

	return len(filterNonKiUsers(activeUsers))
}
func (u *UserSer) UpdateStatistic(user repository.RegisteredUser) (*mongo.UpdateResult, error) {
	return u.UserRepository.UpdateGameStatisticValues(user)
}
func (u *UserSer) UpdatePosition(id string, position string) (int64, error) {
	mongoResult, err := u.UserRepository.UpdatePosition(id, position)

	updateResult := mongoResult.UpsertedCount + mongoResult.ModifiedCount + mongoResult.MatchedCount

	return updateResult, err
}

func (u *UserSer) ShufflePositionsForActiveUsers() {
	activeUsers, _ := u.UserRepository.GetAllActive()
	shuffledPositions := funk.Shuffle([3]string{"1", "2", "3"}).([]string)

	activeNonKiUsers := filterNonKiUsers(activeUsers)

	for index, user := range activeNonKiUsers {
		u.UpdatePosition(user.Id.Hex(), shuffledPositions[index])
	}
}

func (u *UserSer) UpdateState(id string, state string) (int64, error) {
	mongoResult, err := u.UserRepository.UpdateState(id, state)

	updateResult := mongoResult.UpsertedCount + mongoResult.ModifiedCount + mongoResult.MatchedCount

	return updateResult, err
}
func (u *UserSer) GetByGameId(gameId primitive.ObjectID) ([]repository.RegisteredUser, error) {
	return u.UserRepository.GetByGameId(gameId)
}

func (u *UserSer) MapUserIdsToPositions(formParameters map[string][]string) ([]Tuple, error) {
	userIdsToStates, err := mapToTuples(formParameters, "id", "position")

	if err != nil {
		log.Printf("can not map form parameters to id:position tuples %s\n", err)
		return nil, err
	}

	return userIdsToStates, nil
}
func (u *UserSer) MapUserIdsToStates(formParameters map[string][]string) ([]Tuple, error) {
	userIdsToStates, err := mapToTuples(formParameters, "id", "state")

	if err != nil {
		log.Printf("can not map form parameters to id:state tuples %s\n", err)
		return nil, err
	}

	return userIdsToStates, nil
}
func (u *UserSer) FilterNewUserIdsForActivation(userIdsToStates []Tuple) []string {

	activeUsers, _ := u.GetAllActive()
	return filterNewActiveUsers(userIdsToStates, activeUsers)
}
func (u *UserSer) toUserEntry(userEntity *repository.RegisteredUser) *UserEntry {

	userEntry := UserEntry{
		Id:                  userEntity.Id.Hex(),
		RegisteredTimestamp: userEntity.RegistrationTimestamp.Format(repository.GermanDateTimeFormat),
		DisplayName:         userEntity.DisplayName,
		Email:               userEntity.Email,
		PlayedGames:         userEntity.PlayedGames,
		Pos:                 userEntity.Pos,
		State:               userEntity.State,
	}

	if userEntity.LastTimePlayed != nil {
		lastTimePlayed := (*userEntity.LastTimePlayed).Format(repository.GermanDateTimeFormat)
		userEntry.LastTimePlayed = &lastTimePlayed
	}

	return &userEntry
}

func (d *DashboardUser) toUserEntity() *repository.RegisteredUser {
	return &repository.RegisteredUser{
		AcceptNewsletter:   d.AcceptNewsletter,
		AcceptNotification: d.AcceptNotification,
		DisplayName:        d.DisplayName,
		Email:              d.Email,
		FirstName:          d.FirstName,
		LastName:           d.LastName,
	}
}
