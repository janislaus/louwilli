package repository

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

func existsCollection(database *mongo.Database, collectionName string) (bool, *mongo.Collection) {
	cursor, err := database.ListCollections(context.Background(), bson.D{{"name", collectionName}})

	if err != nil {
		log.Fatal(fmt.Sprintf("can not list collections: %s", err))
		return false, nil
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
			return false, nil
		}

		return true, database.Collection(collectionName)
	}

	return false, nil
}
