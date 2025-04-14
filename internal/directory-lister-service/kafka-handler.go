package directorylisterservice

import (
	"context"
	"fmt"
	"ftp-scanner_try2/config"
	ftpclient "ftp-scanner_try2/internal/ftp"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"log"
	"time"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

type KafkaHandlerInterface interface {
	Start(ctx context.Context, config config.DirectoryListerKafkaConfig)
}

type KafkaHandler struct {
	service  DirectoryListerService
	consumer kafka.KafkaDirectoryConsumerInterface
	logger   *zap.Logger
}

func NewKafkaHandler(service DirectoryListerService,
	consumer kafka.KafkaDirectoryConsumerInterface,
	logger *zap.Logger,
) KafkaHandlerInterface {
	return &KafkaHandler{
		service:  service,
		consumer: consumer,
		logger:   logger,
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
			h.logger.Info("Directory-lister-service: KafkaHandler - Остановка обработки сообщений")
			// log.Println("Остановка обработки сообщений")
			return
		default:
			msg, err := h.consumer.ReadMessage(ctx)
			if err != nil {
				h.logger.Error("Directory-lister-service: KafkaHandler - Ошибка чтения сообщения из Kafka", zap.Error(err))
				// log.Println("Ошибка чтения сообщения:", err)
				ReadErrors.Inc()
				continue
			}
			ReceivedMessages.Inc()

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
					h.logger,
				)

				if err != nil {
					h.logger.Error("Directory-lister-service: KafkaHandler - Ошибка подключения к FTP", zap.Error(err))
					continue
				}

				currentFTPClient = ftpClient
				currentParams = &msg.FTPConnection
				FtpReconnections.Inc()
				h.logger.Info("Directory-lister-service: KafkaHandler - Установлено новое FTP-соединение")

				// log.Println("directory-lister kafka-handler Start: Установлено новое FTP-соединение")
			}

			// Проверка активности существующего соединения
			if err := currentFTPClient.CheckConnection(); err != nil {
				h.logger.Warn("Directory-lister-service: KafkaHandler - FTP-соединение неактивно, переподключаемся...")
				currentFTPClient.Close()

				ftpClient, err := ftpclient.NewFTPClient(
					fmt.Sprintf("%s:%d", msg.FTPConnection.Server, msg.FTPConnection.Port),
					msg.FTPConnection.Username,
					msg.FTPConnection.Password,
					h.logger,
				)

				if err != nil {
					h.logger.Error("Directory-lister-service: KafkaHandler - Ошибка подключения к FTP при повторном подключении", zap.Error(err))
					// log.Println("directory-lister kafka-handler Start: Ошибка подключения к FTP:", err)
					continue
				}

				currentFTPClient = ftpClient
				currentParams = &msg.FTPConnection
				FtpReconnections.Inc()
				h.logger.Info("Directory-lister-service: KafkaHandler - Установлено новое FTP-соединение")
			}

			// log.Printf("directory-lister kafka-handler Start: Обработка сообщения %v\n", msg)
			h.logger.Info("Directory-lister-service: KafkaHandler - Обработка Kafka-сообщения", zap.Any("message", msg))
			// Обработка сообщения
			span, ctxWithSpan := opentracing.StartSpanFromContext(ctx, "HandleKafkaMessage")
			startTime := time.Now()
			if err := h.service.ProcessDirectory(ctxWithSpan, msg, currentFTPClient); err != nil {
				log.Println("directory-lister kafka-handler Start: Ошибка обработки директории:", err)
			}
			duration := time.Since(startTime).Seconds()
			ProcessingDuration.Observe(duration)
			span.Finish()
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
