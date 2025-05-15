package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User represents an application user and their config.
type User struct {
	ID        string    `bson:"_id,omitempty"`
	Username  string    `bson:"username"`
	Password  string    `bson:"passwordHash"`
	Endpoints []string  `bson:"endpoints"`
	CreatedAt time.Time `bson:"createdAt"`
}

// UserStore provides CRUD operations on users.
type UserStore struct {
	col *mongo.Collection
}

// NewUserStore initializes a store with given Mongo URI.
func NewUserStore(uri string) (*UserStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	col := client.Database("platform").Collection("users")
	return &UserStore{col: col}, nil
}
