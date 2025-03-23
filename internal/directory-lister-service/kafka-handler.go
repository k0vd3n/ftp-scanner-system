package directorylisterservice

import (
	"context"
	"fmt"
	"ftp-scanner_try2/config"
	ftpclient "ftp-scanner_try2/internal/ftp"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"log"
)

type KafkaHandlerInterface interface {
	Start(ctx context.Context, config config.DirectoryListerKafkaConfig)
}

type KafkaHandler struct {
	service  DirectoryListerService
	consumer kafka.KafkaDirectoryConsumerInterface
}

func NewKafkaHandler(service DirectoryListerService, consumer kafka.KafkaDirectoryConsumerInterface) KafkaHandlerInterface {
	return &KafkaHandler{
		service:  service,
		consumer: consumer,
	}
}

func (h *KafkaHandler) Start(ctx context.Context, config config.DirectoryListerKafkaConfig) {
	var (
		currentFTPClient ftpclient.FtpClientInterface
		currentParams    *models.FTPConnection
	)

	defer func() {
		if currentFTPClient != nil {
			currentFTPClient.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Остановка обработки сообщений")
			return
		default:
			msg, err := h.consumer.ReadMessage(ctx)
			if err != nil {
				log.Println("Ошибка чтения сообщения:", err)
				continue
			}

			// Проверяем необходимость нового подключения
			needNewConnection := currentFTPClient == nil ||
				!compareFTPParams(currentParams, &msg.FTPConnection)

			if needNewConnection {
				if currentFTPClient != nil {
					currentFTPClient.Close()
				}

				ftpClient, err := ftpclient.NewFTPClient(
					fmt.Sprintf("%s:%d", msg.FTPConnection.Server, msg.FTPConnection.Port),
					msg.FTPConnection.Username,
					msg.FTPConnection.Password,
				)

				if err != nil {
					log.Println("Ошибка подключения к FTP:", err)
					continue
				}

				currentFTPClient = ftpClient
				currentParams = &msg.FTPConnection
				log.Println("Установлено новое FTP-соединение")
			}

			// Проверка активности существующего соединения
			if err := currentFTPClient.CheckConnection(); err != nil {
				log.Println("Соединение неактивно, переподключаемся...")
				currentFTPClient.Close()
				currentFTPClient = nil
				continue
			}

			// Обработка сообщения
			if err := h.service.ProcessDirectory(msg, currentFTPClient); err != nil {
				log.Println("Ошибка обработки директории:", err)
				currentFTPClient.Close()
				currentFTPClient = nil
			}
		}
	}
}

func compareFTPParams(a *models.FTPConnection, b *models.FTPConnection) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Server == b.Server &&
		a.Port == b.Port &&
		a.Username == b.Username &&
		a.Password == b.Password
}
