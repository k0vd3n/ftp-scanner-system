package kafka

import (
	"context"
	"encoding/json"
	"ftp-scanner_try2/internal/models"
	"log"

	"github.com/segmentio/kafka-go"
)

type FileConsumer struct {
	Reader *kafka.Reader
}

func NewFileConsumer(brokers []string, topic, groupID string) KafkaFileConsumerInterface {
	return &FileConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

func (c *FileConsumer) ReadMessage(ctx context.Context) (*models.FileScanMessage, error) {
	log.Printf("File-consumer: Чтение сообщений из Kafka...") 
	msg, err := c.Reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var scanMsg models.FileScanMessage
	err = json.Unmarshal(msg.Value, &scanMsg)
	if err != nil {
		log.Printf("File-consumer: Ошибка при разборе сообщения: %v", err)
		return nil, err
	}

	log.Printf("File-consumer: Получено сообщение: %+v\n", scanMsg)
	return &scanMsg, nil
}

func (c *FileConsumer) CloseReader() error {
	log.Println("File-consumer: Закрытие Kafka-консьюмера...")
	return c.Reader.Close()
}
