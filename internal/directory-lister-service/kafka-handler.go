package directorylisterservice

import (
	"context"
	"fmt"
	"ftp-scanner_try2/config"
	ftpclient "ftp-scanner_try2/internal/ftp"
	"ftp-scanner_try2/internal/kafka"
	"log"
)

type KafkaHandlerInterface interface {
	Start(ctx context.Context, config config.DirectoryListerConfig)
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

func (h *KafkaHandler) Start(ctx context.Context, config config.DirectoryListerConfig) {
	host := ""
	user := ""
	password := ""
	port := 0
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
			if scanMsg.FTPConnection.Server != host || scanMsg.FTPConnection.Port != port || scanMsg.FTPConnection.Username != user || scanMsg.FTPConnection.Password != password {
				ftpClient, err = ftpclient.NewFTPClient(fmt.Sprintf("%s:%d", scanMsg.FTPConnection.Server, scanMsg.FTPConnection.Port), scanMsg.FTPConnection.Username, scanMsg.FTPConnection.Password)
				if err != nil {
					log.Fatal("Ошибка подключения к FTP:", err)
				}
				host = scanMsg.FTPConnection.Server
				user = scanMsg.FTPConnection.Username
				password = scanMsg.FTPConnection.Password
				port = scanMsg.FTPConnection.Port
			}

			err = h.service.ProcessDirectory(scanMsg, ftpClient)
			if err != nil {
				log.Println("Ошибка обработки директории:", err)
			}
		}
	}
}
