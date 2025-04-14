package counterreducerservice

import (
	"context"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"ftp-scanner_try2/internal/mongodb"
	"log"
	"time"
)

type CounterReducerService interface {
	ReduceMessages(messages []models.CountMessage) (reducedMessage []models.CountMessage, totalNumber int)
	InsertReducedCounters(ctx context.Context, counts []models.CountMessage) error
	Start(ctx context.Context)
}

type CounterReducer struct {
	repo     mongodb.CounterReducerRepository
	consumer kafka.KafkaCounterReducerConsumerInterface
	cfg      config.UnifiedConfig
}

func NewCounterReducerService(
	repo mongodb.CounterReducerRepository,
	consumer kafka.KafkaCounterReducerConsumerInterface,
	cfg config.UnifiedConfig,
) CounterReducerService {
	return &CounterReducer{
		repo:     repo,
		consumer: consumer,
		cfg:      cfg,
	}
}

func (c *CounterReducer) ReduceMessages(messages []models.CountMessage) (reducedMessage []models.CountMessage, totalNumber int) {
	log.Println("counter-reducer-service reduce-messages: Начало редьюса сообщений...")
	reduced := make(map[string]int)
	totalNumber = 0

	log.Printf("counter-reducer-service reduce-messages: Количество сообщений: %d", len(messages))
	for _, msg := range messages {
		totalNumber += msg.Number
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

	return result, totalNumber
}

func (c *CounterReducer) InsertReducedCounters(ctx context.Context, counts []models.CountMessage) error {
	log.Println("counter-reducer-service insert-reduced-counters: Начало сохранения редьюсированных данных...")
	return c.repo.InsertReducedCounters(ctx, counts)
}

func (c *CounterReducer) Start(ctx context.Context) {
	for {
		messages, err := c.consumer.ReadMessages(
			ctx,
			c.cfg.CounterReducer.Kafka.BatchSize,
			time.Duration(c.cfg.CounterReducer.Kafka.Duration)*time.Second,
		)

		if err != nil {
			log.Printf("counter-reducer-service main: Ошибка чтения сообщений: %v", err)
			continue
		}
		log.Printf("counter-reducer-service main: Получено %d сообщений", len(messages))
		if len(messages) != 0 {

			ReceivedMessages.Add(float64(len(messages)))
			
			startTime := time.Now()
			startReduce := time.Now()
			
			log.Println("counter-reducer-service main: Редьюс сообщений...")
			reduced, totalNumber := c.ReduceMessages(messages)
			timeReduce := time.Since(startReduce).Seconds()
			ProcessingReduce.Observe(timeReduce)
			ReceivedNumbers.Add(float64(totalNumber))
			
			log.Println("counter-reducer-service main: Вставка редьюсированных данных в MongoDB...")
			ReducedMessages.Add(float64(len(reduced)))
			
			startInsert := time.Now()
			if err := c.repo.InsertReducedCounters(ctx, reduced); err != nil {
				log.Printf("counter-reducer-service main: Ошибка вставки данных в MongoDB: %v", err)
				MongoInsertionErrors.Inc()
			} else {
				log.Printf("counter-reducer-service main: Данные вставлены в MongoDB")
				insertTime := time.Since(startInsert).Seconds()
				ProcessingInsert.Observe(insertTime)
			}
			duration := time.Since(startTime).Seconds()
			ProcessingDuration.Observe(duration)
			log.Printf("counter-reducer-service main: Сообщения отправлены в Mongodb")
		}
	}
}
