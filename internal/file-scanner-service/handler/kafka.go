package handler

import (
	"context"
	"fmt"
	filescannerservice "ftp-scanner_try2/internal/file-scanner-service"
	"ftp-scanner_try2/internal/file-scanner-service/service"
	ftpclient "ftp-scanner_try2/internal/ftp"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"log"
)

type KafkaHandlerInterface interface {
	Start(ctx context.Context)
}

type KafkaHandler struct {
	service  service.FileScannerService
	consumer kafka.KafkaFileConsumerInterface
}

func NewKafkaHandler(service service.FileScannerService, consumer kafka.KafkaFileConsumerInterface) KafkaHandlerInterface {
	return &KafkaHandler{
		service:  service,
		consumer: consumer,
	}
}

func (h *KafkaHandler) Start(ctx context.Context) {
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
			log.Println("file-scanner kafka-handler Start: Остановка обработки сообщений")
			return
		default:
			scanMsg, err := h.consumer.ReadMessage(ctx)
			if err != nil {
				log.Println("file-scanner kafka-handler Start: Ошибка чтения сообщения:", err)
				continue
			}

			// Проверяем необходимость нового подключения
			needNewConnection := currentFTPClient == nil ||
				!compareFTPParams(currentParams, &scanMsg.FTPConnection)

			if needNewConnection {
				if currentFTPClient != nil {
					currentFTPClient.Close()
				}

				ftpClient, err := ftpclient.NewFTPClient(
					fmt.Sprintf("%s:%d", scanMsg.FTPConnection.Server, scanMsg.FTPConnection.Port),
					scanMsg.FTPConnection.Username,
					scanMsg.FTPConnection.Password,
				)

				if err != nil {
					log.Println("file-scanner kafka-handler Start: Ошибка подключения к FTP:", err)
					continue // Пропускаем сообщение при ошибке подключения
				}

				currentFTPClient = ftpClient
				currentParams = &scanMsg.FTPConnection
				log.Println("file-scanner kafka-handler Start: Установлено новое FTP-соединение")
			}

			// Проверка активности соединения
			if err := currentFTPClient.CheckConnection(); err != nil {
				log.Println("file-scanner kafka-handler Start: FTP-Соединение неактивно, переподключаемся...")
				currentFTPClient.Close()

				ftpClient, err := ftpclient.NewFTPClient(
					fmt.Sprintf("%s:%d", scanMsg.FTPConnection.Server, scanMsg.FTPConnection.Port),
					scanMsg.FTPConnection.Username,
					scanMsg.FTPConnection.Password,
				)

				if err != nil {
					log.Println("file-scanner kafka-handler Start: Ошибка подключения к FTP:", err)
					continue
				}

				currentFTPClient = ftpClient
				currentParams = &scanMsg.FTPConnection
				filescannerservice.FtpReconnections.Inc()
				log.Println("file-scanner kafka-handler Start: Установлено новое FTP-соединение")
			}

			// Обработка файла
			if err := h.service.ProcessFile(scanMsg, currentFTPClient); err != nil {
				log.Println("file-scanner kafka-handler Start: Ошибка обработки файла:", err)
				// Сбрасываем соединение при ошибке
				currentFTPClient.Close()
				currentFTPClient = nil
				currentParams = nil
			}
		}
	}
}

// Вспомогательная функция для сравнения параметров подключения
func compareFTPParams(a *models.FTPConnection, b *models.FTPConnection) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Server == b.Server &&
		a.Port == b.Port &&
		a.Username == b.Username &&
		a.Password == b.Password
}
