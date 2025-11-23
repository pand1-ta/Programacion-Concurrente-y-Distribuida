package database

import (
	"context"
	"time"

	"sdr/api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewMongoClient(uri, dbName, _col string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	db := client.Database(dbName)
	return &MongoClient{Client: client, DB: db}, nil
}

func (m *MongoClient) SeedMoviesIfEmpty(movies []models.Movie) error {
	coll := m.DB.Collection("movies")
	ctx := context.Background()

	cnt, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return err
	}

	if cnt > 0 {
		return nil
	}

	docs := make([]interface{}, 0, len(movies))
	for _, mv := range movies {
		docs = append(docs, bson.M{
			"movieId": mv.MovieID,
			"title":   mv.Title,
		})
	}

	_, err = coll.InsertMany(ctx, docs)
	return err
}

func (m *MongoClient) SeedUsersIfEmpty(userIdxToOrig map[int]string) error {
	coll := m.DB.Collection("users")
	ctx := context.Background()
	cnt, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	docs := make([]interface{}, 0, len(userIdxToOrig))
	for idx, orig := range userIdxToOrig {
		docs = append(docs, bson.M{"userIndex": idx, "userId": orig})
	}
	if len(docs) == 0 {
		return nil
	}
	_, err = coll.InsertMany(ctx, docs)
	return err
}

func (m *MongoClient) SaveRecommendation(rec interface{}) error {
	coll := m.DB.Collection("history")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := coll.InsertOne(ctx, rec)
	return err
}
