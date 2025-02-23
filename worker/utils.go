package worker

import (
	"log"
	"time"
	"scraper/db"
	"strconv"
)

func RemoveDuplicates(submissions []db.Submission) []db.Submission {
	latest := make(map[string]db.Submission)
	order := []string{} // To maintain insertion order

	for _, sub := range submissions {
		if existing, found := latest[sub.TitleSlug]; !found {
			latest[sub.TitleSlug] = sub
			order = append(order, sub.TitleSlug) // Store order of first appearance
		} else {
			existingTime, _ := strconv.Atoi(existing.Timestamp)
			newTime, _ := strconv.Atoi(sub.Timestamp)
			if newTime > existingTime {
				latest[sub.TitleSlug] = sub
			}
		}
	}

	// Maintain order while collecting results
	result := make([]db.Submission, 0, len(order))
	for _, slug := range order {
		result = append(result, latest[slug])
	}

	return result
}

func RunPeriodically(interval time.Duration, f func()) {
	for {
		begin := time.Now()
		f()
		elapsed := time.Since(begin)
		if elapsed < interval {
			log.Printf("Sleeping for %v until next sync", interval - elapsed)
			time.Sleep(interval - elapsed)
		}
	}
}