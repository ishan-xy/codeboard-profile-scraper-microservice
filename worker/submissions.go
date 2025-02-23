package worker

import (
	"encoding/json"
	"log"
	"scraper/db"
	"time"

	go_utils "github.com/ItsMeSamey/go_utils"
)

// > Go Implementation
// const maxConcurrent = 100

// func FetchAllUserSubmissions() {
//     log.Println("Starting fetching all user submissions")
//     usernames := GetCachedUsernames()
// 	log.Println("Fetched", len(usernames), "usernames")
//     wg := sync.WaitGroup{}
//     sem := make(chan struct{}, maxConcurrent)

//     for _, username := range usernames {
//         wg.Add(1)
//         go func(u string) {
//             sem <- struct{}{} // Acquire semaphore
//             defer func() {
//                 <-sem // Release semaphore
//                 wg.Done()
//             }()

//             variables := map[string]interface{}{
//                 "username": u,
//                 "limit":    20,
//             }

//             s, err := SendQuery(SOLVED_QUESTIONS_QUERY, variables)
//             if err != nil {
//                 log.Println("Error fetching submissions for", u, ":", err)
//                 return
//             }

//             sjson, err := json.Marshal(s)
//             if err != nil {
//                 log.Println("Error marshaling response:", err)
//                 return
//             }
//             var response db.SubmissionResponse
//             if err := json.Unmarshal(sjson, &response); err != nil {
// 				log.Println("Error parsing JSON for", u, ":", err)
//                 return
//             }

//             us := RemoveDuplicates(response.Data.RecentAcSubmissionList)
//             storeSubmissions(u, us)
//         }(username)
//     }

//     wg.Wait() // Wait for all goroutines to complete
//     log.Println("Finished fetching all user submissions")
// }

// func storeSubmissions(username string, submissions []db.Submission) {
// 	// print for now
// 	log.Println("Storing submissions for", username)
// 	for _, sub := range submissions {
// 		log.Println(username, sub.Title)
// 	}
// }

type RequestData struct {
	Usernames []string `json:"usernames"`
	Query     string   `json:"query"`
}
type ResponseData struct {
	Results     map[string][]db.Submission `json:"results"`
	ProcessTime float64                 `json:"process_time"`
}

type UserSubmissions map[string][]db.Submission

func FetchAllUserSubmissions() {
	usernames := GetCachedUsernames()
	requestData := RequestData{
		Usernames: usernames,
		Query:     SOLVED_QUESTIONS_QUERY,
	}

	requestJSON, err := json.Marshal(requestData)
	if err != nil {
		log.Println("Error marshaling request data:", err)
		return
	}
	natsConn, err  := ConnectNats()
	if err != nil {
		log.Println("Error connecting to NATS:", go_utils.WithStack(err))
		return
	}
	defer natsConn.Close()
	msg, err := natsConn.Request("usernames", requestJSON, 30*time.Second)
	if err != nil {
		log.Println(go_utils.WithStack(err))
	}
	
	var response ResponseData
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		log.Println("Error parsing JSON:", go_utils.WithStack(err))
		return
	}
	log.Println("Received response:")
	for username, submissions := range response.Results {
		log.Printf("%s: %d submissions\n", username, len(submissions))
		for _, submission := range submissions {
			log.Printf("  - %s (slug: %s, timestamp: %s)\n", submission.Title, submission.TitleSlug, submission.Timestamp)
		}
	}
	log.Printf("Processing time: %.2f seconds\n", response.ProcessTime)
}
