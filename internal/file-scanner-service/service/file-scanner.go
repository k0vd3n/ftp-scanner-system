package service

import (
	"context"
	"fmt"
	"ftp-scanner_try2/config"
	filescannerservice "ftp-scanner_try2/internal/file-scanner-service"
	"ftp-scanner_try2/internal/file-scanner-service/scanner"
	ftpclient "ftp-scanner_try2/internal/ftp"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type FileScannerService interface {
	ProcessFile(scanMsg *models.FileScanMessage, ftpClient ftpclient.FtpClientInterface, counterOfAllMessages int) error
	ReturnMessage(scanMsg *models.FileScanMessage) error
}

type fileScannerService struct {
	scanResultProducer kafka.KafkaScanResultPoducerInterface
	producer           kafka.KafkaPoducerInterface
	config             config.FileScannerConfig
	scannerMap         map[string]scanner.FileScanner
	logger             *zap.Logger
}

func NewFileScannerService(
	scanResultProducer kafka.KafkaScanResultPoducerInterface,
	counterProducer kafka.KafkaPoducerInterface,
	config config.FileScannerConfig,
	scannerMap map[string]scanner.FileScanner,
	logger *zap.Logger,
) FileScannerService {
	return &fileScannerService{
		scanResultProducer: scanResultProducer,
		producer:           counterProducer,
		config:             config,
		scannerMap:         scannerMap,
		logger:             logger,
	}
}

func (s *fileScannerService) ProcessFile(scanMsg *models.FileScanMessage, ftpClient ftpclient.FtpClientInterface, counterOfAllMessages int) error {
	/*
		ctx := context.Background()
		s.logger.Info("Starting file processing",
			zap.String("file", scanMsg.FilePath),
			zap.String("scan_id", scanMsg.ScanID))

		// Подготовка локальной директории
		localDir := filepath.Join(s.config.KafkaScanResultProducer.FileScanDownloadPath, scanMsg.ScanID)
		if err := s.createLocalDir(localDir); err != nil {
			s.logger.Error("file-scanner-service service: ProcessFile: Ошибка при создании локальной директории",
				zap.String("file", scanMsg.FilePath),
				zap.String("scan_id", scanMsg.ScanID))
			return err
		}
		s.logger.Info("file-scanner-service service: ProcessFile: локальная директория создана",
			zap.String("file", scanMsg.FilePath),
			zap.String("scan_id", scanMsg.ScanID))

		// Скачивание файла с ретраями
		if err := s.retryDownloadFile(ctx, ftpClient, scanMsg.FilePath, localDir, s.config.MaxRetries, time.Duration(s.config.TimeoutSeconds)*time.Second); err != nil {
			s.logger.Error("file-scanner-service service: ProcessFile: Ошибка при скачивании файла",
				zap.String("file", scanMsg.FilePath),
				zap.String("scan_id", scanMsg.ScanID))
			return s.handleDownloadError(ctx, scanMsg, err)
		}
		s.logger.Info("file-scanner-service service: ProcessFile: файл успешно скачан",
			zap.String("file", scanMsg.FilePath),
			zap.String("scan_id", scanMsg.ScanID))

		// Сканирование файла
		result, err := s.scanFile(scanMsg, localDir)
		if err != nil {
			s.logger.Error("file-scanner-service service: ProcessFile: Ошибка при сканировании файла",
				zap.String("file", scanMsg.FilePath),
				zap.String("scan_id", scanMsg.ScanID))
			return err
		}
		s.logger.Info("file-scanner-service service: ProcessFile: файл успешно сканирован",
			zap.String("file", scanMsg.FilePath),
			zap.String("scan_id", scanMsg.ScanID))

		// Отправка результатов
		if err := s.sendScanResults(ctx, scanMsg, result); err != nil {
			s.logger.Error("file-scanner-service service: ProcessFile: Ошибка при отправке результатов сканирования",
				zap.String("file", scanMsg.FilePath),
				zap.String("scan_id", scanMsg.ScanID))
			return err
		}
		s.logger.Info("file-scanner-service service: ProcessFile: результаты успешно отправлены",
			zap.String("file", scanMsg.FilePath),
			zap.String("scan_id", scanMsg.ScanID))

		// Удаление файла и отправка счетчиков
		if err := s.cleanupAndSendCompletedFileCounter(ctx, scanMsg, localDir); err != nil {
			s.logger.Error("file-scanner-service service: ProcessFile: Ошибка при удалении файла и отправке счетчиков",
				zap.String("file", scanMsg.FilePath),
				zap.String("scan_id", scanMsg.ScanID))
			return err
		}
		s.logger.Info("file-scanner-service service: ProcessFile: файл успешно удален",
			zap.String("file", scanMsg.FilePath),
			zap.String("scan_id", scanMsg.ScanID))
		return nil
	*/

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		s.logger.Error("file-scanner-service service: ProcessFile: PANIC RECOVERED",
	// 			zap.String("file", scanMsg.FilePath),
	// 			zap.String("scan_id", scanMsg.ScanID),
	// 			zap.Any("reason", r))
	// 		if err := s.ReturnMessage(scanMsg); err != nil {
	// 			s.logger.Error("file-scanner-service service: ProcessFile: Ошибка при возврате сообщения обратно в топик",
	// 				zap.String("file", scanMsg.FilePath),
	// 				zap.String("scan_id", scanMsg.ScanID),
	// 				zap.String("reason", "PANIC RECOVERED"),
	// 				zap.Error(err))
	// 		}

	// 	}
	// }()
	s.logger.Info("file-scanner-service service: Сканируем файл",
		// zap.Int("kafkaMsg", counterOfAllMessages),
		zap.String("scanID", scanMsg.ScanID),
		zap.String("file", scanMsg.FilePath),
	)

	localDir := filepath.Join(s.config.KafkaScanResultProducer.FileScanDownloadPath, scanMsg.ScanID)
	perm, err := strconv.ParseUint(s.config.KafkaScanResultProducer.Permision, 8, 32)

	if err != nil {
		return fmt.Errorf("ERROR: file-scanner-service service: kafkaMsg: %d, Ошибка при парсинге прав доступа: %v", counterOfAllMessages, err)
	}

	fileMode := os.FileMode(perm)
	s.logger.Info("file-scanner-service service: Создание директории",
		// zap.Int("kafkaMsg", counterOfAllMessages),
		zap.String("scanID", scanMsg.ScanID),
		zap.String("dir", localDir),
		zap.Uint64("perm", perm),
	)

	if err := os.MkdirAll(localDir, fileMode); err != nil {
		s.logger.Error("file-scanner-service service: Ошибка при создании директории",
			// zap.Int("kafkaMsg", counterOfAllMessages),
			zap.String("scanID", scanMsg.ScanID),
			zap.String("dir", localDir),
			zap.Error(err),
		)
		return err
	}

	// Запускаем скачивание в отдельной горутине
	s.logger.Info("file-scanner-service service: Запускаем скачивание в отдельной горутине",
		// zap.Int("kafkaMsg", counterOfAllMessages),
		zap.String("scanID", scanMsg.ScanID),
	)

	s.logger.Info("file-scanner-service service: Готовимся к скачиванию с ретраями",
		// zap.Int("kafkaMsg", counterOfAllMessages),
		zap.String("scanID", scanMsg.ScanID),
		zap.Int("maxRetries", s.config.MaxRetries),
		zap.Int("timeoutSec", s.config.TimeoutSeconds),
	)

	var downloadErr error
	for attempt := range s.config.MaxRetries {

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.TimeoutSeconds)*time.Second)

		resultChan := make(chan error, 1)

		s.logger.Info("file-scanner-service service: Попытка скачивания файла",
			// zap.Int("kafkaMsg", counterOfAllMessages),
			zap.String("scanID", scanMsg.ScanID),
			zap.String("file", scanMsg.FilePath),
			zap.Int("attempt", attempt),
		)

		go func() {
			err := ftpClient.DownloadFile(scanMsg.FilePath, localDir)
			resultChan <- err
			close(resultChan)
		}()

		select {
		case <-ctx.Done():
			downloadErr = ctx.Err()
			s.logger.Warn("file-scanner-service service: Таймаут при скачивании",
				// zap.Int("kafkaMsg", counterOfAllMessages),
				zap.String("scanID", scanMsg.ScanID),
				zap.String("file", scanMsg.FilePath),
				zap.Int("attempt", attempt),
				zap.Error(downloadErr))
		case err := <-resultChan:
			downloadErr = err
			if downloadErr != nil {
				s.logger.Warn("file-scanner-service service: Ошибка при скачивании",
					// zap.Int("kafkaMsg", counterOfAllMessages),
					zap.String("scanID", scanMsg.ScanID),
					zap.String("file", scanMsg.FilePath),
					zap.Int("attempt", attempt),
					zap.Error(downloadErr),
				)
			} else {
				s.logger.Info("file-scanner-service service: Файл успешно скачан",
					// zap.Int("kafkaMsg", counterOfAllMessages),
					zap.String("scanID", scanMsg.ScanID),
					zap.String("file", scanMsg.FilePath),
					zap.Int("attempt", attempt),
				)
			}
		}

		cancel()

		if downloadErr == nil {
			break
		}

		if attempt == s.config.MaxRetries {
			s.logger.Error("file-scanner-service service: Все попытки скачивания исчерпаны",
				zap.String("scanID", scanMsg.ScanID),
				zap.Int("kafkaMsg", counterOfAllMessages),
				zap.String("file", scanMsg.FilePath),
				zap.Int("attempts", s.config.MaxRetries),
				zap.Error(downloadErr),
			)
			s.logger.Info("file-scanner-service: Возвращаем файл обратно в Kafka",
				zap.String("scanID", scanMsg.ScanID),
				// zap.Int("kafkaMsg", counterOfAllMessages),
				zap.String("file", scanMsg.FilePath),
			)
			startTime := time.Now()
			if err := s.producer.SendMessage(s.config.KafkaConsumer.ConsumerTopic, models.FileScanMessage{
				ScanID:        scanMsg.ScanID,
				FilePath:      scanMsg.FilePath,
				ScanType:      scanMsg.ScanType,
				FTPConnection: scanMsg.FTPConnection,
			}); err != nil {
				s.logger.Error("file-scanner-service: Ошибка при возврате файла в топик Kafka",
					// zap.Int("kafkaMsg", counterOfAllMessages),
					zap.String("scanID", scanMsg.ScanID),
					zap.String("file", scanMsg.FilePath),
					zap.Error(err),
				)
				filescannerservice.ReturnErrorDuration.Observe(time.Since(startTime).Seconds())
				filescannerservice.ErrorCounter.Inc()
			} else {
				filescannerservice.ReturningFilesDuration.Observe(time.Since(startTime).Seconds())
				filescannerservice.ReturnedFiles.Inc()
				s.logger.Info("file-scanner-service: Файл успешно возвращен в топик Kafka",
					zap.String("scanID", scanMsg.ScanID),
					// zap.Int("kafkaMsg", counterOfAllMessages),
					zap.String("file", scanMsg.FilePath),
				)
			}
			return downloadErr
		}
	}
	/*








		resultChan := make(chan error, 1)
		// Создаём контекст с таймаутом (например, 30 секунд; значение меняется в конфиге)
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.TimeoutSeconds)*time.Second)
		defer cancel()

		go func() {
			// startDownload := time.Now()
			err := ftpClient.DownloadFile(scanMsg.FilePath, localDir)
			// downloadTime := time.Since(startDownload).Seconds()
			// filescannerservice.DownloadDuration.Observe(downloadTime)
			// filescannerservice.DownloadedFiles.Inc()

			resultChan <- err
			close(resultChan)
		}()

		// Ожидаем завершения скачивания или истечения контекста
		select {
		case <-ctx.Done():

			s.logger.Warn("file-scanner-service: Таймаут при скачивании файла",
				zap.Int("kafkaMsg", counterOfAllMessages),
				zap.String("file", scanMsg.FilePath),
			)
			s.logger.Info("file-scanner-service: Возвращаем файл обратно в Kafka",
				zap.Int("kafkaMsg", counterOfAllMessages),
				zap.String("file", scanMsg.FilePath),
			)
			startTime := time.Now()
			if err := s.producer.SendMessage(s.config.KafkaConsumer.ConsumerTopic, models.FileScanMessage{
				ScanID:        scanMsg.ScanID,
				FilePath:      scanMsg.FilePath,
				ScanType:      scanMsg.ScanType,
				FTPConnection: scanMsg.FTPConnection,
			}); err != nil {
				s.logger.Error("file-scanner-service: Ошибка при возврате файла в топик Kafka",
					zap.Int("kafkaMsg", counterOfAllMessages),
					zap.String("file", scanMsg.FilePath),
					zap.Error(err),
				)
				filescannerservice.ReturnErrorDuration.Observe(time.Since(startTime).Seconds())
				filescannerservice.ErrorCounter.Inc()
			} else {
				filescannerservice.ReturningFilesDuration.Observe(time.Since(startTime).Seconds())
				filescannerservice.ReturnedFiles.Inc()
				s.logger.Info("file-scanner-service: Файл успешно возвращен в топик Kafka",
					zap.Int("kafkaMsg", counterOfAllMessages),
					zap.String("file", scanMsg.FilePath),
				)
			}
			return ctx.Err()

		case err := <-resultChan:

			if err != nil {
				s.logger.Error("file-scanner-service: Ошибка при скачивании файла",
					zap.Int("kafkaMsg", counterOfAllMessages),
					zap.String("file", scanMsg.FilePath),
					zap.Error(err),
				)
				return err
			}
		}

		s.logger.Info("file-scanner-service: Файл успешно скачан",
			zap.Int("kafkaMsg", counterOfAllMessages),
			zap.String("file", scanMsg.FilePath),
		)




	*/
	s.logger.Info("file-scanner-service service: Запускаем сканирование",
		zap.String("scanID", scanMsg.ScanID),
		// zap.Int("kafkaMsg", counterOfAllMessages),
	)

	startScan := time.Now()
	scanner, exists := s.scannerMap[scanMsg.ScanType]

	if !exists {
		s.logger.Error("file-scanner-service service: Тип сканирования не поддерживается",
			zap.String("scanID", scanMsg.ScanID),
			// zap.Int("kafkaMsg", counterOfAllMessages),
			zap.String("scanType", scanMsg.ScanType),
		)
		return nil
	}

	result, err := scanner.Scan(filepath.Join(localDir, filepath.Base(scanMsg.FilePath)))

	if err != nil {
		s.logger.Error("file-scanner-service service: Ошибка при сканировании файла",
			zap.String("scanID", scanMsg.ScanID),
			// zap.Int("kafkaMsg", counterOfAllMessages),
			zap.String("file", scanMsg.FilePath),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("file-scanner-service service: Результат сканирования",
		zap.String("scanID", scanMsg.ScanID),
		// zap.Int("kafkaMsg", counterOfAllMessages),
		zap.String("result", result),
	)
	filescannerservice.ScannedFiles.Inc()
	scanTime := time.Since(startScan).Seconds()
	filescannerservice.ScanDuration.Observe(scanTime)
	s.logger.Info("file-scanner-service service: Удаляем файл после сканирования",
		// zap.Int("kafkaMsg", counterOfAllMessages),
		zap.String("scanID", scanMsg.ScanID),
		zap.String("file", scanMsg.FilePath),
	)
	// Теперь можно удалить файл:

	if err := os.Remove(filepath.Join(localDir, filepath.Base(scanMsg.FilePath))); err != nil {
		s.logger.Error("file-scanner-service service: Не удалось удалить файл",
			// zap.Int("kafkaMsg", counterOfAllMessages),
			zap.String("scanID", scanMsg.ScanID),
			zap.String("file", scanMsg.FilePath),
			zap.Error(err),
		)
	} else {
		s.logger.Info("file-scanner-service service: Файл успешно удалён после сканирования",
			zap.String("scanID", scanMsg.ScanID),
			// zap.Int("kafkaMsg", counterOfAllMessages),
			zap.String("file", scanMsg.FilePath),
		)
	}

	s.logger.Info("file-scanner-service service: Отправляем результат в топик сканирования",
		zap.String("scanID", scanMsg.ScanID),
		// zap.Int("kafkaMsg", counterOfAllMessages),
	)

	// Отправка результата сканирования в Kafka
	startTime := time.Now()

	if err := s.scanResultProducer.SendMessage(models.ScanResultMessage{
		ScanID:   scanMsg.ScanID,
		FilePath: scanMsg.FilePath,
		ScanType: scanMsg.ScanType,
		Result:   result,
	}); err != nil {

		s.logger.Error("file-scanner-service service: Ошибка при отправке результата сканирования в Kafka",
			// zap.Int("kafkaMsg", counterOfAllMessages),
			zap.String("scanID", scanMsg.ScanID),
			zap.Error(err),
		)
		filescannerservice.ErrorCounter.Inc()
		filescannerservice.ErrorSendingScanFilesResultDuration.Observe(time.Since(startTime).Seconds())
		// return err
	}

	filescannerservice.SendingScanFilesResultsDuration.Observe(time.Since(startTime).Seconds())
	filescannerservice.SendedScanFilesResults.Inc()
	s.logger.Info("file-scanner-service service: Отправка результатов сканирования завершена",
		// zap.Int("kafkaMsg", counterOfAllMessages),
		zap.String("scanID", scanMsg.ScanID),
	)

	countMessage := models.CountMessage{
		ScanID: scanMsg.ScanID,
		Number: 1,
	}

	s.logger.Info("file-scanner-service service: Отправляем количество завершенных файлов",
		// zap.Int("kafkaMsg", counterOfAllMessages),
		zap.String("scanID", scanMsg.ScanID),
		zap.Any("countMessage", countMessage),
	)

	// Отправка 1, как количество завершенных файлов
	startTime = time.Now()

	if err := s.producer.SendMessage(s.config.KafkaCompletedFilesCountProducer.CompletedFilesCountTopic, countMessage); err != nil {
		s.logger.Error("file-scanner-service service: Ошибка при отправке количества отсканированных файлов в Kafka",
			// zap.Int("kafkaMsg", counterOfAllMessages),
			zap.String("scanID", scanMsg.ScanID),
			zap.Error(err),
		)
		filescannerservice.ErrorSendingCompletedFilesDuration.Observe(time.Since(startTime).Seconds())
		filescannerservice.ErrorCounter.Inc()

		// return err
	} else {
		s.logger.Info("file-scanner-service service: Отправка количества завершенных файлов завершена",
			// zap.Int("kafkaMsg", counterOfAllMessages),
			zap.String("scanID", scanMsg.ScanID),
		)
		filescannerservice.SendingCompletedFilesDuration.Observe(time.Since(startTime).Seconds())
		filescannerservice.SendedCompletedFiles.Inc()
	}

	s.logger.Info("file-scanner-service service: Отправка количества завершенных файлов завершена",
		// zap.Int("kafkaMsg", counterOfAllMessages),
		zap.String("scanID", scanMsg.ScanID),
	)

	return nil

}

func (s *fileScannerService) ReturnMessage(scanMsg *models.FileScanMessage) error {
	// нужно добавить ретраи на возврат в топик
	for attempt := 0; attempt < s.config.MaxRetries; attempt++ {
		if err := s.producer.SendMessage(s.config.KafkaConsumer.ConsumerTopic, scanMsg); err == nil {
			s.logger.Info("file-scanner-service service: Успешный возврат сообщения об ошибке в Kafka",
				zap.String("scanID", scanMsg.ScanID),
			)
			return nil
		} else {
			s.logger.Error("file-scanner-service service: Ошибка при возврате сообщения обратно в Kafka",
				zap.String("scanID", scanMsg.ScanID),
				zap.Error(err),
			)
		}
	}
	return fmt.Errorf("file-scanner-service service: Ошибка при возврате сообщения обратно в Kafka после %d попыток", s.config.MaxRetries)
}

/*
func (s *fileScannerService) retryDownloadFile(
	ctx context.Context,
	ftpClient ftpclient.FtpClientInterface,
	remotePath string,
	localDir string,
	maxRetries int,
	timeoutPerAttempt time.Duration,
) error {
	for attempt := range maxRetries {
		resultChan := make(chan error, 1)
		dlCtx, cancel := context.WithTimeout(ctx, timeoutPerAttempt)

		s.logger.Info("File-scanner-service: retryDownloadFile: Начало скачивания файла",
			zap.String("file", remotePath),
			zap.Int("attempt", attempt+1))

		go func() {
			start := time.Now()
			err := ftpClient.DownloadFile(remotePath, localDir)
			filescannerservice.DownloadDuration.Observe(time.Since(start).Seconds())
			resultChan <- err
		}()

		select {
		case err := <-resultChan:
			cancel()
			if err == nil {
				filescannerservice.DownloadedFiles.Inc()
				return nil
			}
			s.logger.Error("File-scanner-service: retryDownloadFile: Ошибка при скачивании файла",
				zap.String("file", remotePath),
				zap.Int("attempt", attempt+1),
				zap.Error(err))
		case <-dlCtx.Done():
			cancel()
			s.logger.Warn("File-scanner-service: retryDownloadFile: Таймаут при скачивании файла",
				zap.String("file", remotePath),
				zap.Int("attempt", attempt+1))
		}
	}
	return fmt.Errorf("directory-lister-service: retryDownloadFile: ошибка скачивания файла '%s' после %d попыток", remotePath, maxRetries)
}

func (s *fileScannerService) retryKafkaSend(
	ctx context.Context,
	topic string,
	msg interface{},
	maxRetries int,
	timeoutPerAttempt time.Duration,
	logContext string,
) error {
	for attempt := range maxRetries {
		resultChan := make(chan error, 1)
		kafkaCtx, cancel := context.WithTimeout(ctx, timeoutPerAttempt)

		go func() {
			err := s.producer.SendMessage(topic, msg)
			resultChan <- err
		}()

		select {
		case err := <-resultChan:
			cancel()
			if err == nil {
				return nil
			}
			s.logger.Error("File-scanner-service: retryKafkaSend: Kafka send failed",
				zap.String("topic", topic),
				zap.Int("attempt", attempt+1),
				zap.String("context", logContext),
				zap.Error(err))
		case <-kafkaCtx.Done():
			cancel()
			s.logger.Warn("File-scanner-service: retryKafkaSend: Kafka send timeout",
				zap.String("topic", topic),
				zap.Int("attempt", attempt+1),
				zap.String("context", logContext))
		}
	}
	return fmt.Errorf("file-scanner-service: retryKafkaSend: failed to send to Kafka topic %s after %d attempts", topic, maxRetries)
}

func (s *fileScannerService) retryKafkaSendScanResult(
	ctx context.Context,
	msg models.ScanResultMessage,
	maxRetries int,
	timeoutPerAttempt time.Duration,
	logContext string,
) error {
	for attempt := range maxRetries {
		resultChan := make(chan error, 1)
		kafkaCtx, cancel := context.WithTimeout(ctx, timeoutPerAttempt)

		go func() {
			err := s.scanResultProducer.SendMessage(msg)
			resultChan <- err
		}()

		select {
		case err := <-resultChan:
			cancel()
			if err == nil {
				return nil
			}
			s.logger.Error("File-scanner-service: retryKafkaSend: Ошибка при отправке в Kafka",
				// zap.String("topic", topic),
				zap.Int("attempt", attempt+1),
				zap.String("context", logContext),
				zap.Error(err))
		case <-kafkaCtx.Done():
			cancel()
			s.logger.Warn("File-scanner-service: retryKafkaSend: Ошибка, таймаут при отправке в Kafka",
				// zap.String("topic", topic),
				zap.Int("attempt", attempt+1),
				zap.String("context", logContext))
		}
	}
	return fmt.Errorf("file-scanner-service: retryKafkaSend: Ошибка при отправке результата сканирования в Kafka после %d попыток", maxRetries)
}

func (s *fileScannerService) createLocalDir(localDir string) error {
	perm, err := strconv.ParseUint(s.config.KafkaScanResultProducer.Permision, 8, 32)
	if err != nil {
		return fmt.Errorf("FileScannerService: createLocalDir: недопустимое значение permision: %w", err)
	}

	if err := os.MkdirAll(localDir, os.FileMode(perm)); err != nil {
		return fmt.Errorf("FileScannerService: createLocalDir: не удалось создать директорию: %w", err)
	}
	return nil
}

func (s *fileScannerService) handleDownloadError(ctx context.Context, scanMsg *models.FileScanMessage, err error) error {
	msg := models.FileScanMessage{
		ScanID:        scanMsg.ScanID,
		FilePath:      scanMsg.FilePath,
		ScanType:      scanMsg.ScanType,
		FTPConnection: scanMsg.FTPConnection,
	}

	logContext := fmt.Sprintf("FileScannerService: handleDownloadError: возвращение файла %s обратно в топик", scanMsg.FilePath)
	if kafkaErr := s.retryKafkaSend(
		ctx,
		s.config.KafkaConsumer.ConsumerTopic,
		msg,
		s.config.MaxRetries,
		time.Duration(s.config.TimeoutSeconds)*time.Second,
		logContext,
	); kafkaErr != nil {
		filescannerservice.ErrorCounter.Inc()
		return fmt.Errorf("file-scanner-service service: handleDownloadError: Ошибка при возвращении файла %s обратно в топик: %w", scanMsg.FilePath, kafkaErr)
	}

	filescannerservice.ReturnedFiles.Inc()
	return fmt.Errorf("file-scanner-service service: handleDownloadError: Ошибка при скачивании файла %s: %w", scanMsg.FilePath, err)
}

func (s *fileScannerService) scanFile(scanMsg *models.FileScanMessage, localDir string) (string, error) {
	scanner, exists := s.scannerMap[scanMsg.ScanType]
	if !exists {
		return "", fmt.Errorf("file-scanner-service service: scanFile: неизвестный тип сканирования: %s", scanMsg.ScanType)
	}

	startScan := time.Now()
	result, err := scanner.Scan(filepath.Join(localDir, filepath.Base(scanMsg.FilePath)))
	if err != nil {
		return "", fmt.Errorf("file-scanner-service service: scanFile: ошибка сканирования: %w", err)
	}

	filescannerservice.ScannedFiles.Inc()
	filescannerservice.ScanDuration.Observe(time.Since(startScan).Seconds())
	return result, nil
}

func (s *fileScannerService) sendScanResults(ctx context.Context, scanMsg *models.FileScanMessage, result string) error {
	startTime := time.Now()
	msg := models.ScanResultMessage{
		ScanID:   scanMsg.ScanID,
		FilePath: scanMsg.FilePath,
		ScanType: scanMsg.ScanType,
		Result:   result,
	}

	logContext := fmt.Sprintf("file-scanner-service service: sendScanResults: отправка результатов сканирования %s", scanMsg.FilePath)
	if err := s.retryKafkaSendScanResult(
		ctx,
		msg,
		s.config.MaxRetries,
		time.Duration(s.config.TimeoutSeconds)*time.Second,
		logContext,
	); err != nil {
		filescannerservice.ErrorCounter.Inc()
		return fmt.Errorf("file-scanner-service service: sendScanResults: ошибка при отправке результатов сканирования: %w", err)
	}

	filescannerservice.SendedScanFilesResults.Inc()
	filescannerservice.SendingScanFilesResultsDuration.Observe(time.Since(startTime).Seconds())
	return nil
}

func (s *fileScannerService) cleanupAndSendCompletedFileCounter(ctx context.Context, scanMsg *models.FileScanMessage, localDir string) error {
	// Удаление файла
	filePath := filepath.Join(localDir, filepath.Base(scanMsg.FilePath))
	if err := os.Remove(filePath); err != nil {
		s.logger.Warn("file-scanner-service service: cleanupAndSendCounters: удаление файла",
			zap.String("file", filePath),
			zap.Error(err))
	}

	// Отправка счетчика завершенных файлов
	countMsg := models.CountMessage{
		ScanID: scanMsg.ScanID,
		Number: 1,
	}

	logContext := fmt.Sprintf("file-scanner-service service: cleanupAndSendCounters: отправка счетчика завершенных файлов %s", scanMsg.FilePath)
	if err := s.retryKafkaSend(
		ctx,
		s.config.KafkaCompletedFilesCountProducer.CompletedFilesCountTopic,
		countMsg,
		s.config.MaxRetries,
		time.Duration(s.config.TimeoutSeconds)*time.Second,
		logContext,
	); err != nil {
		filescannerservice.ErrorCounter.Inc()
		return fmt.Errorf("file-scanner-service service: cleanupAndSendCounters: ошибка при отправке счетчика завершенных файлов: %w", err)
	}

	filescannerservice.SendedCompletedFiles.Inc()
	return nil
}
*/
