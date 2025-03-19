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
	log.Printf("Directory-consumer: Чтение сообщений из Kafka...")
	msg, err := c.Reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var scanMsg models.DirectoryScanMessage
	err = json.Unmarshal(msg.Value, &scanMsg)
	if err != nil {
		return nil, err
	}

	log.Printf("Directory-consumer: Получено сообщение: %+v\n", scanMsg)
	return &scanMsg, nil
}

func (c *DirectoryConsumer) CloseReader() error {
	log.Println("Directory-consumer: Закрытие Kafka-консьюмера...")
	return c.Reader.Close()
}
