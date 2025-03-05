package counterreducerservice

import (
	"context"
	"ftp-scanner_try2/internal/models"
	"ftp-scanner_try2/internal/mongodb"
)

type CounterReducerService interface {
	ReduceMessages(messages []models.CountMessage) []models.CountMessage
	InsertReducedCounters(ctx context.Context, counts []models.CountMessage) error
}

type CounterReducer struct {
	repo mongodb.CounterReducerRepository
}

func NewCounterReducerService(repo mongodb.CounterReducerRepository) CounterReducerService {
	return &CounterReducer{repo: repo}
}

func (c *CounterReducer) ReduceMessages(messages []models.CountMessage) []models.CountMessage {
	reduced := make(map[string]int)

	for _, msg := range messages {
		reduced[msg.ScanID] += msg.Number
	}

	var result []models.CountMessage
	for scanID, sum := range reduced {
		result = append(result, models.CountMessage{
			ScanID: scanID,
			Number: sum,
		})
	}

	return result
}

func (c *CounterReducer) InsertReducedCounters(ctx context.Context, counts []models.CountMessage) error {
	return c.repo.InsertReducedCounters(ctx, counts)
}


