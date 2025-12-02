package models

type Movie struct {
	MovieID string `json:"movieId" bson:"movieId"`
	Title   string `json:"title" bson:"title"`
	Genre   string `json:"genre" bson:"genre"`
}
