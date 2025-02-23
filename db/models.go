package db

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Collection[T any] struct {
	*mongo.Collection
}
type UserPublicProfile struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	LeetcodeUsername string             `bson:"leetcode_username"`
	ProfilePic       string             `json:"profile_pic" bson:"profile_pic"`
	QuestionCount    int                `bson:"question_count"`
	Ranking          int                `bson:"ranking"`
	TotalQuestions   int                `bson:"total_questions"`
	Solved           []Solved           `bson:"solved"`
}

type Solved struct {
	Timestamp int64    `bson:"timestamp"`
	Question  Question `bson:"question"`
}

type Question struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	QuestionID int                `json:"q_id" bson:"q_id"`
	Difficulty DifficultyLevel    `json:"difficulty" bson:"difficulty"`
	Title      string             `json:"title" bson:"title"`
	TitleSlug  string             `json:"titleSlug" bson:"title_slug"`
}
