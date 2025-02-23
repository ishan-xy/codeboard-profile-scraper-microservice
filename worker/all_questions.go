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

func FetchAllQuestions() (map[string]db.Question, error) {
	variables := map[string]interface{}{
		"categorySlug": "all-code-essentials",
		"skip":         0,
		"limit":        10000,
		"filters":      map[string]interface{}{},
	}
	s, err := SendQuery(ALL_QUESTION_LIST_QUERY, variables)
	if err != nil {
		log.Println(go_utils.WithStack(err))
		return nil, err
	}
	sJson, _ := json.Marshal(s)
	var response db.QuestionListResponse
	if err := json.Unmarshal([]byte(sJson), &response); err != nil {
		log.Println("Error parsing JSON:", go_utils.WithStack(err))
		return nil, err
	}
	questions := response.Data.ProblemsetQuestionList.Questions

	// Convert slice to map with TitleSlug as the key
	questionsMap := make(map[string]db.Question)
	for _, q := range questions {
		questionsMap[q.TitleSlug] = q
	}

	// Store the map in Redis
	questionsJson, _ := json.Marshal(questionsMap)
	err = redisClient.Set(ctx, "all_questions", questionsJson, 24*time.Hour).Err()
	if err != nil {
		log.Println(go_utils.WithStack(err))
		return nil, err
	}	
	log.Println("Fetched", len(questionsMap), "questions")
	return questionsMap, nil
}

func GetCachedQuestions() (map[string]db.Question, error) {
	// Check cache first
	val, err := redisClient.Get(ctx, "all_questions").Result()
	if err == redis.Nil {
		return FetchAllQuestions()
	} else if err != nil {
		log.Println("Redis error:", err)
		return FetchAllQuestions() // Fallback to API
	}

	// Cache hit: deserialize into map
	var questionsMap map[string]db.Question
	err = json.Unmarshal([]byte(val), &questionsMap)
	if err != nil {
		return nil, fmt.Errorf("error decoding cached data: %v", go_utils.WithStack(err))
	}

	log.Println("Retrieved", len(questionsMap), "questions from cache")
	return questionsMap, nil
}