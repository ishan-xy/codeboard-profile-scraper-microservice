package worker

import (
	"log"
	"time"
	"context"
	"scraper/db"
	"encoding/json"
	
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	go_utils "github.com/ItsMeSamey/go_utils"
)

func fetchUsernames() []string {
	var usernames []string
	err := db.UserDB.Distinct(context.TODO(), "leetcode_username", bson.M{}).Decode(&usernames)
	if err != nil {
		log.Println("Error fetching usernames:", go_utils.WithStack(err))
		return nil
	}
	err = SetUsernamesInCache(usernames)
	if err != nil {
		log.Println("Error setting usernames in cache:", go_utils.WithStack(err))
	}
	log.Println("Fetched", len(usernames), "usernames")
	return usernames
}

func GetCachedUsernames() []string {
	// Check cache first
	val, err := redisClient.Get(ctx, "all_usernames").Result()
	if err == redis.Nil {
		return fetchUsernames()
	} else if err != nil {
		log.Println("Redis error:", err)
		return fetchUsernames() // Fallback to API
	}

	// Cache hit: deserialize data
	var usernames []string
	err = json.Unmarshal([]byte(val), &usernames)
	if err != nil {
		log.Println("Error decoding cached data:", go_utils.WithStack(err))
	}
	log.Println("Retrieved", len(usernames), "usernames from cache")
	return usernames
}

func AddUsernameToCache(username string) {
	usernames := GetCachedUsernames()
	usernames = append(usernames, username)
	SetUsernamesInCache(usernames)
}

func SetUsernamesInCache(usernames []string) error{
	usernamesJson, _ := json.Marshal(usernames)
	err := redisClient.Set(ctx, "all_usernames", usernamesJson, 24*time.Hour).Err()
	if err != nil {
		return go_utils.WithStack(err)
	}
	return nil
}