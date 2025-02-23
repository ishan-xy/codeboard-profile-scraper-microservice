package worker

import (
	"log"
	"time"
	"fmt"
	"scraper/db"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	go_utils "github.com/ItsMeSamey/go_utils"
)

func FetchAllQuestions() ([]db.Question , error) {
	variables := map[string]interface{}{
		"categorySlug": "all-code-essentials",
		"skip":         0,
		"limit":        10000,
		"filters":      map[string]interface{}{},
	}
	s, err := SendQuery(ALL_QUESTION_LIST_QUERY, variables)
	if err != nil {
		log.Println(go_utils.WithStack(err))
	}
	sJson, _:= json.Marshal(s)
	var response db.QuestionListResponse
	if err := json.Unmarshal([]byte(sJson), &response); err != nil {
		log.Println("Error parsing JSON:", go_utils.WithStack(err))
		return nil, err
	}
	questions := response.Data.ProblemsetQuestionList.Questions

	// store in redis cache
	questionsJson, _ := json.Marshal(questions)
	err = redisClient.Set(ctx, "all_questions", questionsJson, 24*time.Hour).Err()
	if err != nil {
		log.Println(go_utils.WithStack(err))
	}	
	log.Println("Fetched", len(questions), "questions")
	return questions, nil
}

func GetCachedQuestions() ([]db.Question, error) {
	// Check cache first
	val, err := redisClient.Get(ctx, "all_questions").Result()
	if err == redis.Nil {
		return FetchAllQuestions()
	} else if err != nil {
		log.Println("Redis error:", err)
		return FetchAllQuestions() // Fallback to API
	}

	// Cache hit: deserialize data
	var questions []db.Question
	err = json.Unmarshal([]byte(val), &questions)
	if err != nil {
		return nil, fmt.Errorf("error decoding cached data: %v", go_utils.WithStack(err))
	}

	log.Println("Retrieved", len(questions), "questions from cache")
	return questions, nil
}