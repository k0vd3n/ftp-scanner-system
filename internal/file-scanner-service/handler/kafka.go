package handler

import (
	"context"
	"ftp-scanner_try2/internal/file-scanner-service/service"
	ftpclient "ftp-scanner_try2/internal/ftp"
	"ftp-scanner_try2/internal/kafka"
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
	host := ""
	user := ""
	password := ""
	ftpClient := ftpclient.FtpClientInterface(nil)
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
			if scanMsg.FTPConnection.Server != host || scanMsg.FTPConnection.Username != user || scanMsg.FTPConnection.Password != password {
				ftpClient, err = ftpclient.NewFTPClient(scanMsg.FTPConnection.Server, scanMsg.FTPConnection.Username, scanMsg.FTPConnection.Password)
				if err != nil {
					log.Fatal("Ошибка подключения к FTP:", err)
				}
				host = scanMsg.FTPConnection.Server
				user = scanMsg.FTPConnection.Username
				password = scanMsg.FTPConnection.Password
			}

			err = h.service.ProcessFile(scanMsg, ftpClient)
			if err != nil {
				log.Println("Ошибка обработки файла:", err)
			}
		}
	}
}
