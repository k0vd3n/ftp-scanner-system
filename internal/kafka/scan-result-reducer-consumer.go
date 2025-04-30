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

type ScanResultConsumer struct {
	Reader *kafka.Reader
}

func NewScanResultConsumer(brokers []string, topic, groupID string) KafkaScanResultReducerConsumerInterface {
	return &ScanResultConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
			CommitInterval: 0,
		}),
	}
}

func (c *ScanResultConsumer) ReadMessages(ctx context.Context, batchSize int, duration time.Duration) ([]models.ScanResultMessage, int, error) {
    ctx, cancel := context.WithTimeout(ctx, duration)
    defer cancel()

    var messages []models.ScanResultMessage
    total := 0

    for len(messages) < batchSize {
        // if total >= batchSize {
        //     break
        // }
        msg, err := c.Reader.FetchMessage(ctx)
        if err != nil {
            if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				log.Println("scan-result-reducer-consumer: таймаут, заканчиваем чтение")
                break
            }
            return nil, total, err
        }

        var resultMsg models.ScanResultMessage
        if err := json.Unmarshal(msg.Value, &resultMsg); err != nil {
            log.Printf("scan-result-reducer-consumer: Ошибка unmarshal: %v", err)
            // не прибавляем total — пропускаем
            continue
        }

        messages = append(messages, resultMsg)
        total++
		log.Printf("scan-result-reducer-consumer: Получено сообщение: %v", resultMsg) 

		if err := c.Reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("scan-result-reducer-consumer: Ошибка commit: %v", err)
		}
    }

	log.Printf("scan-result-reducer-consumer: Получено %d сообщений", total)
    return messages, total, nil
}

func (c *ScanResultConsumer) CloseReader() error {
	log.Println("scan-result-reducer-consumer: Закрытие Kafka-консьюмера...")
	return c.Reader.Close()
}
