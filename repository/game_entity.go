package repository

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GameState string

type GameEntity struct {
	Id       primitive.ObjectID `bson:"_id"`
	Duration float64            `bson:"duration"`

	KiName  string `bson:"ki_name"`
	KiCoins int    `bson:"ki_coins"`

	Player1      string `bson:"player1"`
	Player1Coins int    `bson:"player_1_coins"`

	Player2      string `bson:"player_2"`
	Player2Coins int    `bson:"player_2_coins"`

	Player3      string `bson:"player_3"`
	Player3Coins int    `bson:"player_3_coins"`

	State GameState
}

func (c GameState) String() string {
	return string(c)
}

const (
	GameAnnounced GameState = "announced"
	GameReady     GameState = "ready"
	GameActive    GameState = "active"
	GameFinished  GameState = "finished"
)
