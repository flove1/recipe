package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoConnection(connectionString string) (*mongo.Client, error) {
	opt := options.Client().ApplyURI(connectionString)

	err := opt.Validate()
	if err != nil {
		return nil, fmt.Errorf("mongo options validate err: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, fmt.Errorf("mongo connect err: %w", err)
	}

	return client, nil
}
