package db

import (
	"encoding/json"
	"fmt"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DifficultyLevel int

const (
	Easy DifficultyLevel = iota
	Medium
	Hard
)

type Submission struct {
	Timestamp string `json:"timestamp"`
	Title     string `json:"title"`
	TitleSlug string `json:"title_slug"`
}

type SubmissionResponse struct {
	Data struct {
		RecentAcSubmissionList []Submission `json:"recentAcSubmissionList"`
	} `json:"data"`
	Errors interface{} `json:"errors"`
}

type QuestionListResponse struct {
	Data struct {
		ProblemsetQuestionList struct {
			Questions []Question `json:"questions"`
		} `json:"problemsetQuestionList"`
	} `json:"data"`
}

func (q *Question) UnmarshalJSON(data []byte) error {
	type Alias Question // Avoid recursion
	aux := &struct {
		QuestionID interface{} `json:"q_id"`    // Handle string or number
		Difficulty interface{} `json:"difficulty"` // Handle string or number
		*Alias
	}{
		Alias: (*Alias)(q),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	q.ID = primitive.NewObjectID()
	// Parse QuestionID (string or number)
	switch v := aux.QuestionID.(type) {
	case string:
		qid, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid QuestionID (string): %v", v)
		}
		q.QuestionID = qid
	case float64: // JSON numbers are decoded as float64
		q.QuestionID = int(v)
	default:
		return fmt.Errorf("invalid type for QuestionID: %T", v)
	}

	// Parse Difficulty (string or number)
	switch v := aux.Difficulty.(type) {
	case string:
		switch v {
		case "Easy":
			q.Difficulty = Easy
		case "Medium":
			q.Difficulty = Medium
		case "Hard":
			q.Difficulty = Hard
		default:
			return fmt.Errorf("invalid difficulty string: %s", v)
		}
	case float64: // Handle enum values (0, 1, 2) from cache
		switch int(v) {
		case 0:
			q.Difficulty = Easy
		case 1:
			q.Difficulty = Medium
		case 2:
			q.Difficulty = Hard
		default:
			return fmt.Errorf("invalid difficulty value: %d", int(v))
		}
	default:
		return fmt.Errorf("invalid type for difficulty: %T", v)
	}

	return nil
}