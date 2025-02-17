package kafka

import (
	"context"
	"encoding/json"
	"log"

	"ftp-scanner_try2/internal/models"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	Reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

func (c *Consumer) ReadMessage(ctx context.Context) (*models.DirectoryScanMessage, error) {
	msg, err := c.Reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var scanMsg models.DirectoryScanMessage
	err = json.Unmarshal(msg.Value, &scanMsg)
	if err != nil {
		return nil, err
	}

	log.Printf("📥 Получено сообщение: %+v\n", scanMsg)
	return &scanMsg, nil
}

func (c *Consumer) CloseReader() error {
	return c.Reader.Close()
}
