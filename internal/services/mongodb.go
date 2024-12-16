package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

func Services() (*mongo.Client, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file is specified.")
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return nil, errors.New("you must set your 'MONGODB_URI' environment variable")
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %s", err)
	}
	fmt.Println("Connected to MongoDB!")
	return client, nil
}
