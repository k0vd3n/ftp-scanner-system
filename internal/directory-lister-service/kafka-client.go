package directorylisterservice

import (
	"context"
	"ftp-scanner_try2/internal/kafka"
	"log"
)

type KafkaHandler struct {
	service DirectoryListerService
	consumer kafka.KafkaConsumerInterface
}

func NewKafkaHandler(service DirectoryListerService, consumer kafka.KafkaConsumerInterface) *KafkaHandler {
	return &KafkaHandler{
		service:  service,
		consumer: consumer,
	}
}

func (h *KafkaHandler) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Остановка обработки сообщений")
			return
		default:
			scanMsg, err := h.consumer.ReadMessage(ctx)
			if err != nil {
				log.Println("Ошибка чтения сообщения:", err)
				continue
			}

			err = h.service.ProcessDirectory(scanMsg)
			if err != nil {
				log.Println("Ошибка обработки директории:", err)
			}
		}
	}
}