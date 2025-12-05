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
			"genre":   mv.Genre,
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

// Obtener usuarios paginados
func (m *MongoClient) GetUsersPaginated(page, limit int) ([]string, error) {
	coll := m.DB.Collection("users")

	skip := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := &options.FindOptions{
		Skip:  int64Ptr(int64(skip)),
		Limit: int64Ptr(int64(limit)),
	}

	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []string

	for cursor.Next(ctx) {
		var doc struct {
			UserId string `bson:"userId"`
		}
		if err := cursor.Decode(&doc); err == nil {
			users = append(users, doc.UserId)
		}
	}

	return users, nil
}

// Obtener películas paginadas + filtro opcional por género
func (m *MongoClient) GetMoviesPaginated(genre string, page, limit int) ([]models.Movie, error) {
	coll := m.DB.Collection("movies")

	skip := (page - 1) * limit

	filter := bson.M{}
	if genre != "" {
		filter["genre"] = bson.M{"$regex": genre, "$options": "i"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := &options.FindOptions{
		Skip:  int64Ptr(int64(skip)),
		Limit: int64Ptr(int64(limit)),
	}

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var movies []models.Movie

	for cursor.Next(ctx) {
		var mv models.Movie
		if err := cursor.Decode(&mv); err == nil {
			movies = append(movies, mv)
		}
	}

	return movies, nil
}

// Helpers para paginación
func int64Ptr(v int64) *int64 { return &v }
