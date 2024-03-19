package repository

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type UserRepository interface {
	Create(user RegisteredUser) (*mongo.InsertOneResult, error)
	CreateKiUser() (*mongo.InsertOneResult, error)
	Get(email string) (*RegisteredUser, error)
	GetByDisplayName(displayName string) (*RegisteredUser, error)
	GetByGameId(gameId primitive.ObjectID) ([]RegisteredUser, error)
	GetAllActive() ([]RegisteredUser, error)
	GetAll() ([]RegisteredUser, error)
	GetPagedSortedByRegistrationDateWithoutKiUser(page int64, nameFilter string) ([]RegisteredUser, error)
	CountAllWithoutKiUser(nameFilter string) (int64, error)
	UpdateGameStatisticValues(user RegisteredUser) (*mongo.UpdateResult, error)
	UpdateGameRelationship(userId *primitive.ObjectID, gameId *primitive.ObjectID) (*mongo.UpdateResult, error)
	UpdatePosition(id string, position string) (*mongo.UpdateResult, error)
	UpdateState(id string, state string) (*mongo.UpdateResult, error)
	UpdateAllNonKiUsers(state UserState) (*mongo.UpdateResult, error)
	Remove(displayName string) error
}

type UserRepo struct {
	collection *mongo.Collection
}

type UserState string

const (
	UserActive  UserState = "active"
	UserWaiting UserState = "waiting"
)

type RegisteredUser struct {
	Id                    primitive.ObjectID  `bson:"_id"`
	GameId                *primitive.ObjectID `bson:"game_id"`
	RegistrationTimestamp time.Time           `bson:"registration_timestamp"`

	AcceptNewsletter   bool   `bson:"accept_newsletter"`
	AcceptNotification bool   `bson:"accept_notification"`
	DisplayName        string `bson:"display_name"`
	Email              string `bson:"email"`
	FirstName          string `bson:"first_name"`
	LastName           string `bson:"last_name"`

	LastTimePlayed *time.Time `bson:"last_time_played"`
	BestDuration   float64    `bson:"best_duration"`
	GamesWon       int        `bson:"games_won"`
	PlayedGames    int        `bson:"played_games"`
	State          UserState  `bson:"state"`
	Pos            string     `bson:"pos"`

	IsKiUser bool `bson:"is_ki_user"`
}

func NewUserRepo(ctx context.Context, client *mongo.Client, databaseName string) *UserRepo {

	database := client.Database(databaseName)

	exists, existingCollection := existsCollection(database, RegisteredUsersCollection)

	if exists == true {
		log.Printf("user collection exists \n")
		return &UserRepo{collection: existingCollection}
	}

	err := database.CreateCollection(ctx, RegisteredUsersCollection)

	if err != nil {
		log.Fatal(fmt.Sprintf("can not create user collection: %s", err))
	}

	collection := database.Collection(RegisteredUsersCollection)
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{"email", 1}},
		Options: options.Index().SetUnique(true),
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("can not create user collection email index: %s", err))
	}

	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{"display_name", 1}},
		Options: options.Index().SetUnique(true),
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("can not create user collection display_name index: %s", err))
	}

	return &UserRepo{collection: collection}
}

func (config *UserRepo) Remove(displayName string) error {

	ctx := context.Background()

	filter := bson.M{"display_name": displayName}

	_, err := config.collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Printf("deleting user failed %s\n", err)
		return err
	}

	return nil
}

func (config *UserRepo) CreateKiUser() (*mongo.InsertOneResult, error) {

	ctx := context.Background()

	newUser := RegisteredUser{
		Id:                    primitive.NewObjectID(),
		RegistrationTimestamp: time.Now().UTC(),

		AcceptNewsletter:   false,
		AcceptNotification: false,
		DisplayName:        KiName,
		Email:              "",
		FirstName:          "",
		LastName:           "",

		BestDuration: InitialUserDuration,
		GamesWon:     0,
		PlayedGames:  0,
		State:        UserActive,
		Pos:          "-1",

		IsKiUser: true,
	}

	result, err := config.collection.InsertOne(ctx, &newUser)
	if err != nil {
		log.Printf("saving ki user failed %s\n", err)
		return nil, err
	}

	return result, nil
}

func (config *UserRepo) Create(user RegisteredUser) (*mongo.InsertOneResult, error) {

	ctx := context.Background()

	newUser := RegisteredUser{
		Id:                    primitive.NewObjectID(),
		RegistrationTimestamp: time.Now().UTC(),

		AcceptNewsletter:   user.AcceptNewsletter,
		AcceptNotification: user.AcceptNotification,
		DisplayName:        user.DisplayName,
		Email:              user.Email,
		FirstName:          user.FirstName,
		LastName:           user.LastName,

		BestDuration: InitialUserDuration,
		GamesWon:     0,
		PlayedGames:  0,
		State:        UserWaiting,
		Pos:          "1",

		IsKiUser: false,
	}

	result, err := config.collection.InsertOne(ctx, &newUser)
	if err != nil {
		log.Printf("saving new user failed %s\n", err)
		return nil, err
	}

	return result, nil
}

func (config *UserRepo) GetByDisplayName(displayName string) (*RegisteredUser, error) {

	ctx := context.Background()
	var result RegisteredUser

	filter := bson.M{"display_name": displayName}

	mongoSingleResult := config.collection.FindOne(ctx, filter)

	err := mongoSingleResult.Decode(&result)

	if err != nil {
		log.Printf("can not find user by display name: %s %s\n", displayName, err)
		return nil, err
	}

	return &result, nil
}

func (config *UserRepo) Get(email string) (*RegisteredUser, error) {

	ctx := context.Background()
	var result RegisteredUser

	filter := bson.M{"email": email}

	mongoSingleResult := config.collection.FindOne(ctx, filter)

	err := mongoSingleResult.Decode(&result)

	if err != nil {
		log.Printf("can not find user by email: %s %s\n", email, err)
		return nil, err
	}

	return &result, nil
}

func (config *UserRepo) UpdatePosition(id string, position string) (*mongo.UpdateResult, error) {

	ctx := context.Background()

	parsedId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		log.Printf("can not parse a not valid user id %s\n", err)
		return nil, err
	}

	filter := bson.M{"_id": parsedId}
	update := bson.M{"$set": bson.M{"pos": position}}

	mongoSingleResult, err := config.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		log.Printf("some error occured during update user position to %s of user %s: %s\n", position, id, err)
		return nil, err
	}

	return mongoSingleResult, nil
}

func (config *UserRepo) UpdateGameStatisticValues(user RegisteredUser) (*mongo.UpdateResult, error) {

	ctx := context.Background()

	filter := bson.M{"_id": user.Id}
	update := bson.M{"$set": bson.M{
		"best_duration":    user.BestDuration,
		"games_won":        user.GamesWon,
		"played_games":     user.PlayedGames,
		"last_time_played": time.Now(),
	}}

	mongoSingleResult, err := config.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		log.Printf("some error occured during update user game statistics: %s\n", err)
		return nil, err
	}

	return mongoSingleResult, nil
}

func (config *UserRepo) GetByGameId(gameId primitive.ObjectID) ([]RegisteredUser, error) {

	ctx := context.Background()
	registeredUsers := make([]RegisteredUser, 0)

	cursor, err := config.collection.Find(ctx, bson.M{"game_id": gameId})

	if err != nil {
		log.Printf("some error occured during get users by game id %s: %s\n", gameId.Hex(), err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user RegisteredUser
		if err := cursor.Decode(&user); err != nil {
			log.Printf("some error occured during decoding all users received from mongo db: %s\n", err)
			return nil, err
		}
		registeredUsers = append(registeredUsers, user)
	}

	return registeredUsers, nil
}

func (config *UserRepo) UpdateGameRelationship(userId *primitive.ObjectID, gameId *primitive.ObjectID) (*mongo.UpdateResult, error) {

	ctx := context.Background()

	filter := bson.M{"_id": userId}
	update := bson.M{"$set": bson.M{"game_id": gameId}}

	mongoSingleResult, err := config.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		log.Printf("some error occured during update user game relationship: %s\n", err)
		return nil, err
	}

	return mongoSingleResult, nil
}

func (config *UserRepo) UpdateState(id string, state string) (*mongo.UpdateResult, error) {

	ctx := context.Background()

	parsedId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		log.Printf("can not parse a not valid user id %s\n", err)
		return nil, err
	}

	filter := bson.M{"_id": parsedId}
	update := bson.M{"$set": bson.M{"state": state}}

	mongoSingleResult, err := config.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		log.Printf("some error occured during update user state to %s of user %s: %s\n", state, id, err)
		return nil, err
	}

	return mongoSingleResult, nil
}

func (config *UserRepo) GetAllActive() ([]RegisteredUser, error) {

	ctx := context.Background()
	registeredUsers := make([]RegisteredUser, 0)

	cursor, err := config.collection.Find(ctx, bson.M{"state": UserActive})

	if err != nil {
		log.Printf("some error occured during get all active users: %s\n", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user RegisteredUser
		if err := cursor.Decode(&user); err != nil {
			log.Printf("some error occured during decoding all active users received from mongo db: %s\n", err)
			return nil, err
		}
		registeredUsers = append(registeredUsers, user)
	}

	return registeredUsers, nil
}

func (config *UserRepo) CountAllWithoutKiUser(nameFilter string) (int64, error) {

	ctx := context.Background()

	filter := bson.M{
		"display_name": bson.M{"$regex": nameFilter},
		"$and": []bson.M{
			{"is_ki_user": bson.M{"$eq": false}},
		},
	}

	count, err := config.collection.CountDocuments(ctx, filter)

	if err != nil {
		log.Printf("some error occured during get all active users: %s\n", err)
		return -1, err
	}

	return count, nil
}

func (config *UserRepo) GetPagedSortedByRegistrationDateWithoutKiUser(page int64, nameFilter string) ([]RegisteredUser, error) {

	ctx := context.Background()
	registeredUsers := make([]RegisteredUser, 0)

	pageSize := int64(10)
	skip := (page - 1) * pageSize

	filter := bson.M{
		"display_name": bson.M{"$regex": nameFilter},
		"$and": []bson.M{
			{"is_ki_user": bson.M{"$eq": false}},
		},
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "last_time_played", Value: -1}, {Key: "registration_timestamp", Value: -1}}).
		SetSkip(skip).
		SetLimit(pageSize)

	cursor, err := config.collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("some error occured during get paged users: %s\n", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(context.Background()) {
		var user RegisteredUser
		if err := cursor.Decode(&user); err != nil {
			log.Printf("some error occured during decoding paged users received from mongo db: %s\n", err)
			return nil, err
		}
		registeredUsers = append(registeredUsers, user)
	}

	return registeredUsers, nil
}

func (config *UserRepo) GetAll() ([]RegisteredUser, error) {

	ctx := context.Background()
	registeredUsers := make([]RegisteredUser, 0)

	cursor, err := config.collection.Find(ctx, bson.D{})

	if err != nil {
		log.Printf("some error occured during get all users: %s\n", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user RegisteredUser
		if err := cursor.Decode(&user); err != nil {
			log.Printf("some error occured during decoding all users received from mongo db: %s\n", err)
			return nil, err
		}
		registeredUsers = append(registeredUsers, user)
	}

	return registeredUsers, nil
}

func (config *UserRepo) UpdateAllNonKiUsers(state UserState) (*mongo.UpdateResult, error) {

	ctx := context.Background()

	filter := bson.M{"is_ki_user": false}
	update := bson.M{"$set": bson.M{"state": state}}

	mongoSingleResult, err := config.collection.UpdateMany(ctx, filter, update)

	if err != nil {
		log.Printf("some error occured updating all users with state %s %s\n", state, err)
		return nil, err
	}

	return mongoSingleResult, nil
}
