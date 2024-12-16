package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type repository[T any] struct {
	db *mongo.Client
}

func InitializeDB() (*mongo.Client, error) {
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

func (r *repository[T]) Create(model T, collection string) (bool, error) {
	coll := r.db.Database("bhagabad_gita").Collection(collection)
	_, err := coll.InsertOne(context.TODO(), model)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *repository[T]) GetByField(field string, field_value string, collection string) (interface{}, error) {
	coll := r.db.Database("bhagabad_gita").Collection(collection)
	filter := bson.D{{Key: field, Value: field_value}}
	var result map[string]interface{}
	err := coll.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *repository[T]) InsertData(workinforamtion interface{}, collection string) (bool, error) {
	coll := r.db.Database("bhagabad_gita").Collection(collection)
	_, err := coll.InsertOne(context.TODO(), workinforamtion)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *repository[T]) GetData(username string, collection string) (interface{}, error) {
	coll := r.db.Database("bhagabad_gita").Collection(collection)
	filter := bson.D{{Key: "username", Value: username}}
	var result interface{}
	err := coll.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *repository[T]) GetAll(collection string) ([]map[string]interface{}, error) {
	coll := r.db.Database("bhagabad_gita").Collection(collection)
	cursor, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())
	var results []map[string]interface{}
	for cursor.Next(context.TODO()) {
		var result map[string]interface{}
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	fmt.Println("The result is here", results)
	return results, nil
}

func (r *repository[T]) GetAllByField(field string, field_value string, collection string) ([]map[string]interface{}, error) {
	coll := r.db.Database("bhagabad_gita").Collection(collection)
	filter := bson.D{{Key: field, Value: field_value}}
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())
	var results []map[string]interface{}
	for cursor.Next(context.TODO()) {
		var result map[string]interface{}
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
