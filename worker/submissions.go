package worker

import (
	"log"
	"time"
	"strconv"
	"context"
	"scraper/db"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	go_utils "github.com/ItsMeSamey/go_utils"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type RequestData struct {
	Usernames []string `json:"usernames"`
	Query     string   `json:"query"`
}

type ResponseData struct {
	Results     map[string][]db.Submission `json:"results"`
	ProcessTime float64                    `json:"process_time"`
}

type UserSubmissions map[string][]db.Submission

func FetchAllUserSubmissions() error {
	usernames := GetCachedUsernames()
	requestData := RequestData{
		Usernames: usernames,
		Query:     SOLVED_QUESTIONS_QUERY,
	}

	requestJSON, err := json.Marshal(requestData)
	if err != nil {
		return go_utils.WithStack(err)
	}

	msg, err := PublishRequestForData(requestJSON)
	if err != nil {
		return go_utils.WithStack(err)
	}

	var response ResponseData
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		return go_utils.WithStack(err)
	}

	if err := SetSubmissionsInCache(response.Results); err != nil {
		return go_utils.WithStack(err)
	}
	
	return nil
}
func UpdateInDB() {
	subs, err := GetCachedSubmissions()
	if err != nil {
		log.Println(go_utils.WithStack(err))
		return
	}

	quesList, err := GetCachedQuestions()
	if err != nil {
		log.Println(go_utils.WithStack(err))
		return
	}

	collection := db.UserDB
	var operations []mongo.WriteModel
	seen := make(map[string]map[int64]struct{}) // [username][timestamp]struct

	// First pass: collect all existing timestamps from cache
	for username, submissions := range subs {
		seen[username] = make(map[int64]struct{})
		for _, submission := range submissions {
			timestamp, _ := strconv.ParseInt(submission.Timestamp, 10, 64)
			seen[username][timestamp] = struct{}{}
		}
	}

	// Second pass: prepare bulk operations
	for username, submissions := range subs {
		var newSolved []db.Solved
		for _, submission := range submissions {
			q, exists := quesList[submission.TitleSlug]
			if !exists {
				continue
			}

			timestamp, err := strconv.ParseInt(submission.Timestamp, 10, 64)
			if err != nil {
				log.Println(go_utils.WithStack(err))
				continue
			}

			// Skip if we've already seen this timestamp
			if _, exists := seen[username][timestamp]; exists {
				continue
			}

			newSolved = append(newSolved, db.Solved{
				Timestamp: timestamp,
				Question:  q,
			})
		}

		if len(newSolved) > 0 {
			filter := bson.M{"leetcode_username": username}
			update := bson.M{
				"$push": bson.M{
					"solved": bson.M{
						"$each": newSolved,
						"$sort": bson.M{"timestamp": -1}, // Optional: maintain sorted order
					},
				},
			}
			
			operation := mongo.NewUpdateOneModel().
				SetFilter(filter).
				SetUpdate(update).
				SetUpsert(true) // Create user if not exists

			operations = append(operations, operation)
		}
	}

	if len(operations) > 0 {
		bulkOptions := options.BulkWrite().SetOrdered(false)
		result, err := collection.BulkWrite(context.Background(), operations, bulkOptions)
		if err != nil {
			log.Println("Bulk write error:", go_utils.WithStack(err))
			return
		}
		log.Printf("Updated %d users, added %d new solved entries",
			result.ModifiedCount, result.UpsertedCount)
	}
}

func PublishRequestForData(requestJSON []byte) (*nats.Msg, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, go_utils.WithStack(err)
	}
	defer nc.Close()

	msg, err := nc.Request("usernames", requestJSON, 30*time.Second)
	if err != nil {
		return nil, go_utils.WithStack(err)
	}
	return msg, nil
}

func GetCachedSubmissions() (UserSubmissions, error) {
	val, err := redisClient.Get(ctx, "all_submissions").Result()
	if err == redis.Nil {
		if err := FetchAllUserSubmissions(); err != nil {
			return nil, go_utils.WithStack(err)
		}
		return nil, nil
	} else if err != nil {
		return nil, go_utils.WithStack(err)
	}

	var submissions UserSubmissions
	if err := json.Unmarshal([]byte(val), &submissions); err != nil {
		return nil, go_utils.WithStack(err)
	}

	return submissions, nil
}

func SetSubmissionsInCache(submissions UserSubmissions) error {
	submissionsJSON, err := json.Marshal(submissions)
	if err != nil {
		return go_utils.WithStack(err)
	}

	if err := redisClient.Set(ctx, "all_submissions", submissionsJSON, 24*time.Hour).Err(); err != nil {
		return go_utils.WithStack(err)
	}

	return nil
}