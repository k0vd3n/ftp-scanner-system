package kafka

import (
	"context"
	"encoding/json"
	"ftp-scanner_try2/internal/models"
	"log"

	"github.com/segmentio/kafka-go"
)

type CounterConsumer struct {
	Reader *kafka.Reader
}

func NewCounterConsumer(brokers []string, topic, groupID string) KafkaCounterConsumerInterface {
	return &CounterConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

func (c *CounterConsumer) ReadMessages(ctx context.Context, batchSize int) ([]models.CountMessage, error) {
	var messages []models.CountMessage

	for i := 0; i < batchSize; i++ {
		msg, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			return nil, err
		}

		var countMsg models.CountMessage
		if err := json.Unmarshal(msg.Value, &countMsg); err != nil {
			log.Printf("Ошибка при разборе сообщения: %v", err)
			continue
		}

		messages = append(messages, countMsg)
	}

	return messages, nil
}

func (c *CounterConsumer) CloseReader() error {
	return c.Reader.Close()
}
