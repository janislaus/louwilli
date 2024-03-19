package repository

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type GameRepository interface {
	CreateGame(gameMembers []RegisteredUser) (*primitive.ObjectID, error)
	GetCurrent() (*GameEntity, error)
	RemoveGame(gameId string) (*mongo.DeleteResult, error)
	UpdateState(gameId string, state GameState) (*GameEntity, error)
	UpdateDuration(gameId string, duration float64) (*GameEntity, error)
	UpdateCoins(gameId string, playerCoinMarker string, coins int) (*GameEntity, error)
}

type GameRepo struct {
	collection *mongo.Collection
}

func NewGameRepository(ctx context.Context, client *mongo.Client, databaseName string) *GameRepo {

	database := client.Database(databaseName)

	exists, existingCollection := existsCollection(database, GamesCollection)

	if exists == true {
		log.Printf("games collection exists \n")
		return &GameRepo{collection: existingCollection}
	}

	err := database.CreateCollection(ctx, GamesCollection)

	if err != nil {
		log.Fatal(fmt.Sprintf("can not create games collection: %s", err))
	}

	collection := client.Database(databaseName).Collection(GamesCollection)

	return &GameRepo{collection: collection}
}

func (config *GameRepo) CreateGame(gameMembers []RegisteredUser) (*primitive.ObjectID, error) {

	ctx := context.Background()

	existingGame := config.collection.FindOne(ctx, bson.D{})

	if !errors.Is(existingGame.Err(), mongo.ErrNoDocuments) {
		return nil, nil
	}

	var game = GameEntity{
		Id:       primitive.NewObjectID(),
		KiName:   KiName,
		KiCoins:  3,
		State:    GameAnnounced,
		Duration: InitialGameDuration,
	}

	for _, member := range gameMembers {
		switch member.Pos {
		case "1":
			game.Player1 = member.DisplayName
			game.Player1Coins = 3
		case "2":
			game.Player2 = member.DisplayName
			game.Player2Coins = 3
		case "3":
			game.Player3 = member.DisplayName
			game.Player3Coins = 3
		}
	}

	result, err := config.collection.InsertOne(ctx, &game)
	if err != nil {
		log.Printf("saving new game failed %s\n", err)
		return nil, err
	}

	gameId, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Printf("can not cast returned id to mongo primitive object id %s\n", err)
		return nil, err
	}

	return &gameId, nil
}

func (config *GameRepo) GetCurrent() (*GameEntity, error) {

	ctx := context.Background()
	var result GameEntity

	game := config.collection.FindOne(ctx, bson.D{})

	if errors.Is(game.Err(), mongo.ErrNoDocuments) {
		return nil, nil
	}

	err := game.Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (config *GameRepo) RemoveGame(gameId string) (*mongo.DeleteResult, error) {

	ctx := context.Background()

	parsedId, err := primitive.ObjectIDFromHex(gameId)

	if err != nil {
		log.Printf("can not parse a not valid game id %s\n", err)
		return nil, err
	}

	filter := bson.M{"_id": parsedId}

	return config.collection.DeleteOne(ctx, filter)
}

func (config *GameRepo) UpdateState(gameId string, state GameState) (*GameEntity, error) {

	ctx := context.Background()

	parsedId, err := primitive.ObjectIDFromHex(gameId)

	if err != nil {
		log.Printf("can not parse a not valid game id %s\n", err)
		return nil, err
	}

	filter := bson.M{"_id": parsedId}
	update := bson.M{"$set": bson.M{"state": state}}

	_, err = config.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		log.Printf("some error occured during update game state to %s of game %s: %s\n", state, gameId, err)
		return nil, err
	}

	game, err := config.GetCurrent()

	if err != nil {
		log.Printf("after updating game state, receiving of current game failed: %s\n", err)
		return nil, err
	}

	return game, nil
}

func (config *GameRepo) UpdateDuration(gameId string, duration float64) (*GameEntity, error) {

	ctx := context.Background()

	parsedId, err := primitive.ObjectIDFromHex(gameId)

	if err != nil {
		log.Printf("can not parse a not valid game id %s\n", err)
		return nil, err
	}

	filter := bson.M{"_id": parsedId}
	update := bson.M{"$set": bson.M{"duration": duration}}

	_, err = config.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		log.Printf("some error occured during update game duration to %f of game %s: %s\n", duration, gameId, err)
		return nil, err
	}

	game, err := config.GetCurrent()

	if err != nil {
		log.Printf("after updating game duration, receiving of current game failed: %s\n", err)
		return nil, err
	}

	return game, nil
}

func (config *GameRepo) UpdateCoins(gameId string, playerCoinMarker string, coins int) (*GameEntity, error) {

	ctx := context.Background()

	parsedId, err := primitive.ObjectIDFromHex(gameId)

	if err != nil {
		log.Printf("can not parse a not valid game id %s\n", err)
		return nil, err
	}

	filter := bson.M{"_id": parsedId}
	update := bson.M{"$set": bson.M{fmt.Sprintf("%s", playerCoinMarker): coins}}

	_, err = config.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		log.Printf("some error occured during update player coins of game %s: %s\n", gameId, err)
		return nil, err
	}

	game, err := config.GetCurrent()

	if err != nil {
		log.Printf("after updating game coins, receiving of current game failed: %s\n", err)
		return nil, err
	}

	return game, nil
}
