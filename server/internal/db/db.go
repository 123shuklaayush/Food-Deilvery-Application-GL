package db

import (
    "context"
    "os"

    "go.mongodb.org/mongo-driver/mongo"
)

type Collections struct {
    Users       *mongo.Collection
    Restaurants *mongo.Collection
    Orders      *mongo.Collection
}

func NewCollections(client *mongo.Client) *Collections {
    databaseName := os.Getenv("MONGODB_DATABASE")
    if databaseName == "" {
        databaseName = "food_delivery"
    }
    database := client.Database(databaseName)
    return &Collections{
        Users:       database.Collection("users"),
        Restaurants: database.Collection("restaurants"),
        Orders:      database.Collection("orders"),
    }
}

func Ping(ctx context.Context, client *mongo.Client) error {
    return client.Ping(ctx, nil)
}



