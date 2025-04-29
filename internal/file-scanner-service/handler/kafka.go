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

	"go.uber.org/zap"
)

type KafkaHandlerInterface interface {
	Start(ctx context.Context)
}

type KafkaHandler struct {
	service    service.FileScannerService
	consumer   kafka.KafkaFileConsumerInterface
	logger     *zap.Logger
	cancelFunc context.CancelFunc
	// context    context.Context
	currentMsg *models.FileScanMessage // для хранения текущего сообщения
}

func NewKafkaHandler(service service.FileScannerService,
	consumer kafka.KafkaFileConsumerInterface,
	logger *zap.Logger,
	cancelFunc context.CancelFunc,
	// context context.Context,
) KafkaHandlerInterface {
	return &KafkaHandler{
		service:    service,
		consumer:   consumer,
		logger:     logger,
		cancelFunc: cancelFunc,
		// context:  context,
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
		if r := recover(); r != nil {
			h.logger.Error("file-scanner kafka-handler Start: PANIC RECOVERED", zap.Any("reason", r))
			if h.currentMsg != nil {
				if err := h.service.ReturnMessage(h.currentMsg); err != nil {
					h.logger.Error("file-scanner kafka-handler Start: Ошибка при возврате сообщения обратно в топик",
						zap.String("scanID", h.currentMsg.ScanID),
						zap.String("reason", "PANIC RECOVERED"),
						zap.Error(err))
				} else {
					h.logger.Info("file-scanner kafka-handler Start: Сообщение возвращено обратно в топик",
						zap.String("scanID", h.currentMsg.ScanID),
						zap.String("reason", "PANIC RECOVERED"))
					filescannerservice.ReturnedFiles.Inc()
				}
			} else {
				h.logger.Error("file-scanner kafka-handler Start: текущее сообщение не задано",
					zap.String("reason", "PANIC RECOVERED"))
			}
			h.cancelFunc()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("file-scanner kafka-handler Start: Остановка обработки сообщений")
			return
		default:
			//================
			// if err := h.service.ProcessFile(h.currentMsg, currentFTPClient, 1); err != nil {
			// 	log.Println("file-scanner kafka-handler Start: Ошибка обработки сообщения:", err)
			// 	continue
			// }

			//================
			counterOfAllMessages := 1
			scanMsg, err := h.consumer.ReadMessage(ctx)
			if err != nil {
				log.Println("file-scanner kafka-handler Start: Ошибка чтения сообщения:", err)
				continue
			}

			// h.logger.Info("file-scanner kafka-handler Start: значение метрики полученных сообщений", zap.String("value", filescannerservice.ReceivedMessages.Desc().String()))
			filescannerservice.ReceivedMessages.Inc()
			// metric := &dto.Metric{}
			// if err := filescannerservice.ReceivedMessages.Write(metric); err == nil {
			// 	log.Printf("Увеличено ReceivedMessages, новое значение: %f", metric.Counter.GetValue())
			// }

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
					h.logger,
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
					h.logger,
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
			if err := h.service.ProcessFile(scanMsg, currentFTPClient, counterOfAllMessages); err != nil {
				counterOfAllMessages++
				log.Println("file-scanner kafka-handler Start: Ошибка обработки файла:", err)
				// Сбрасываем соединение при ошибке
				// currentFTPClient.Close()
				// currentFTPClient = nil
				// currentParams = nil
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
