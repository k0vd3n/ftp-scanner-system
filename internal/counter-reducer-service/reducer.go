package counterreducerservice

import (
	"context"
	"ftp-scanner_try2/internal/models"
	"ftp-scanner_try2/internal/mongodb"
	"log"
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
	log.Println("counter-reducer-service reduce-messages: Начало редьюса сообщений...")
	reduced := make(map[string]int)

	log.Printf("counter-reducer-service reduce-messages: Количество сообщений: %d", len(messages))
	for _, msg := range messages {
		reduced[msg.ScanID] += msg.Number
	}
	log.Printf("counter-reducer-service reduce-messages: Количество уникальных сканов: %d", len(reduced))

	log.Println("counter-reducer-service reduce-messages: Формирование сообщения с редьюсированными данными...")
	var result []models.CountMessage
	for scanID, sum := range reduced {
		result = append(result, models.CountMessage{
			ScanID: scanID,
			Number: sum,
		})
	}
	log.Printf("counter-reducer-service reduce-messages: Результат редьюсирования: %v", result)

	return result
}

func (c *CounterReducer) InsertReducedCounters(ctx context.Context, counts []models.CountMessage) error {
	log.Println("counter-reducer-service insert-reduced-counters: Начало сохранения редьюсированных данных...")
	return c.repo.InsertReducedCounters(ctx, counts)
}


