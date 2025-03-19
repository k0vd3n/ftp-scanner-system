package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"ftp-scanner_try2/internal/models"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type CounterConsumer struct {
	Reader *kafka.Reader
}

func NewCounterConsumer(brokers []string, topic, groupID string) KafkaCounterReducerConsumerInterface {
	return &CounterConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

func (c *CounterConsumer) ReadMessages(ctx context.Context, batchSize int, duration time.Duration) ([]models.CountMessage, error) {
	var messages []models.CountMessage

	// Установка тайм-аута через контекст
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	for i := 0; i < batchSize; i++ {
		select {
		case <-ctx.Done(): // Если время истекло, выходим из цикла
			log.Println("Тайм-аут: прекращаем чтение сообщений")
			return messages, nil
		default:
			msg, err := c.Reader.ReadMessage(ctx)
			if err != nil {
				// Проверяем, не истек ли контекст
				if errors.Is(err, context.DeadlineExceeded) {
					log.Println("Контекст завершен по тайм-ауту")
					return messages, nil
				}
				return nil, err
			}

			var countMsg models.CountMessage
			if err := json.Unmarshal(msg.Value, &countMsg); err != nil {
				log.Printf("Ошибка при разборе сообщения: %v", err)
				continue
			}

			messages = append(messages, countMsg)
		}
	}

	return messages, nil
}

func (c *CounterConsumer) CloseReader() error {
	return c.Reader.Close()
}
