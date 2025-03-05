package kafka

import (
	"context"
	"encoding/json"
	"ftp-scanner_try2/internal/models"
	"log"

	"github.com/segmentio/kafka-go"
)

type ScanResultConsumer struct {
	Reader *kafka.Reader
}

func NewScanResultConsumer(brokers []string, topic, groupID string) KafkaScanResultConsumerInterface {
	return &ScanResultConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

func (c *ScanResultConsumer) ReadMessages(ctx context.Context, batchSize int) ([]models.ScanResultMessage, error) {
	var messages []models.ScanResultMessage

	for i := 0; i < batchSize; i++ {
		msg, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			return nil, err
		}

		var resultMsg models.ScanResultMessage
		if err := json.Unmarshal(msg.Value, &resultMsg); err != nil {
			log.Printf("Ошибка при разборе сообщения: %v", err)
			continue
		}

		messages = append(messages, resultMsg)
	}

	return messages, nil
}

func (c *ScanResultConsumer) CloseReader() error {
	return c.Reader.Close()
}
