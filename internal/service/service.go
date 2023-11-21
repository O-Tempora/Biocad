package service

import "go.mongodb.org/mongo-driver/mongo"

type Service struct {
	Db *mongo.Client
}
