package models

type User struct {
	UserIndex int    `json:"userIndex" bson:"userIndex"`
	UserId    string `json:"userId" bson:"userId"`
}
