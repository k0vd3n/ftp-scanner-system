package kafka

import (
	"context"
	"encoding/json"
	"ftp-scanner_try2/internal/models"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type ScanResultConsumer struct {
	Reader *kafka.Reader
}

func NewScanResultConsumer(brokers []string, topic, groupID string) KafkaScanResultReducerConsumerInterface {
	return &ScanResultConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

func (c *ScanResultConsumer) ReadMessages(ctx context.Context, batchSize int, duration time.Duration) ([]models.ScanResultMessage, error) {
	log.Printf("scan-result-reducer-consumer: Чтение сообщений из Kafka...")
	var messages []models.ScanResultMessage

	// Установка тайм-аута через контекст
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	for i := 0; i < batchSize; i++ {
		select {
		case <-ctx.Done(): // Если время истекло, выходим из цикла
			log.Println("scan-result-reducer-consumer: Тайм-аут: прекращаем чтение сообщений")
			return messages, nil
		default:
			msg, err := c.Reader.ReadMessage(ctx)
			if err != nil {
				return nil, err
			}

			var resultMsg models.ScanResultMessage
			if err := json.Unmarshal(msg.Value, &resultMsg); err != nil {
				log.Printf("scan-result-reducer-consumer: Ошибка при разборе сообщения: %v", err)
				continue
			}

			messages = append(messages, resultMsg)
		}

	}

	log.Printf("scan-result-reducer-consumer: Получено сообщений: %d\n", len(messages))
	return messages, nil
}

func (c *ScanResultConsumer) CloseReader() error {
	log.Println("scan-result-reducer-consumer: Закрытие Kafka-консьюмера...")
	return c.Reader.Close()
}
