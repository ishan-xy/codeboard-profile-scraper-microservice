package worker

import (
	"time"
)
func init() {
	InitRedis()
	InitSubs()
	go RunPeriodically(3 * time.Hour, func() { FetchAllQuestions() })
	go RunPeriodically(1 * time.Hour, func() { fetchUsernames() })
	go RunPeriodically(10 * time.Second, func() { FetchAllUserSubmissions() })

}


