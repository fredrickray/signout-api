package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Client struct {
	client *mongo.Client
	DB     *mongo.Database
}

func Connect(ctx context.Context, uri, dbName string) (*Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	return &Client{client: client, DB: client.Database(dbName)}, nil
}

func (c *Client) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

func (c *Client) EnsureIndexes(ctx context.Context) error {
	users := c.DB.Collection("users")
	_, err := users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]int{"email": 1},
		Options: options.Index().SetUnique(true).SetName("users_email_uidx"),
	})
	if err != nil {
		return fmt.Errorf("users email index: %w", err)
	}

	refresh := c.DB.Collection("refresh_tokens")
	_, err = refresh.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]int{"token_hash": 1},
		Options: options.Index().SetUnique(true).SetName("refresh_token_hash_uidx"),
	})
	if err != nil {
		return fmt.Errorf("refresh token hash index: %w", err)
	}
	_, err = refresh.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]int{"user_id": 1},
		Options: options.Index().SetName("refresh_user_id_idx"),
	})
	if err != nil {
		return fmt.Errorf("refresh user_id index: %w", err)
	}
	return nil
}
