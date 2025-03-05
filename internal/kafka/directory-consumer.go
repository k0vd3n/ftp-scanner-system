package kafka

import (
	"context"
	"encoding/json"
	"log"

	"ftp-scanner_try2/internal/models"

	"github.com/segmentio/kafka-go"
)

type DirectoryConsumer struct {
	Reader *kafka.Reader
}

func NewDirectoryConsumer(brokers []string, topic, groupID string) KafkaDirectoryConsumerInterface {
	return &DirectoryConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

func (c *DirectoryConsumer) ReadMessage(ctx context.Context) (*models.DirectoryScanMessage, error) {
	msg, err := c.Reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var scanMsg models.DirectoryScanMessage
	err = json.Unmarshal(msg.Value, &scanMsg)
	if err != nil {
		return nil, err
	}

	log.Printf(" Получено сообщение: %+v\n", scanMsg)
	return &scanMsg, nil
}

func (c *DirectoryConsumer) CloseReader() error {
	return c.Reader.Close()
}
