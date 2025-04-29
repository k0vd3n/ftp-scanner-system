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
	s.logger.Info("file-scanner-service service: Сканируем файл",
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
		zap.String("scanID", scanMsg.ScanID),
		zap.String("dir", localDir),
		zap.Uint64("perm", perm),
	)

	if err := os.MkdirAll(localDir, fileMode); err != nil {
		s.logger.Error("file-scanner-service service: Ошибка при создании директории",
			zap.String("scanID", scanMsg.ScanID),
			zap.String("dir", localDir),
			zap.Error(err),
		)
		return err
	}

	// Запускаем скачивание в отдельной горутине
	s.logger.Info("file-scanner-service service: Запускаем скачивание в отдельной горутине",
		zap.String("scanID", scanMsg.ScanID),
	)

	s.logger.Info("file-scanner-service service: Готовимся к скачиванию с ретраями",
		zap.String("scanID", scanMsg.ScanID),
		zap.Int("maxRetries", s.config.MaxRetries),
		zap.Int("timeoutSec", s.config.TimeoutSeconds),
	)

	var downloadErr error
	for attempt := range s.config.MaxRetries {

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.TimeoutSeconds)*time.Second)

		resultChan := make(chan error, 1)

		s.logger.Info("file-scanner-service service: Попытка скачивания файла",
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
				zap.String("scanID", scanMsg.ScanID),
				zap.String("file", scanMsg.FilePath),
				zap.Int("attempt", attempt),
				zap.Error(downloadErr))
		case err := <-resultChan:
			downloadErr = err
			if downloadErr != nil {
				s.logger.Warn("file-scanner-service service: Ошибка при скачивании",
					zap.String("scanID", scanMsg.ScanID),
					zap.String("file", scanMsg.FilePath),
					zap.Int("attempt", attempt),
					zap.Error(downloadErr),
				)
			} else {
				s.logger.Info("file-scanner-service service: Файл успешно скачан",
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
					zap.String("file", scanMsg.FilePath),
				)
			}
			return downloadErr
		}
	}
	s.logger.Info("file-scanner-service service: Запускаем сканирование",
		zap.String("scanID", scanMsg.ScanID),
	)

	startScan := time.Now()
	scanner, exists := s.scannerMap[scanMsg.ScanType]

	if !exists {
		s.logger.Error("file-scanner-service service: Тип сканирования не поддерживается",
			zap.String("scanID", scanMsg.ScanID),
			zap.String("scanType", scanMsg.ScanType),
		)
		return nil
	}

	result, err := scanner.Scan(filepath.Join(localDir, filepath.Base(scanMsg.FilePath)))

	if err != nil {
		s.logger.Error("file-scanner-service service: Ошибка при сканировании файла",
			zap.String("scanID", scanMsg.ScanID),
			zap.String("file", scanMsg.FilePath),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("file-scanner-service service: Результат сканирования",
		zap.String("scanID", scanMsg.ScanID),
		zap.String("result", result),
	)
	filescannerservice.ScannedFiles.Inc()
	scanTime := time.Since(startScan).Seconds()
	filescannerservice.ScanDuration.Observe(scanTime)
	s.logger.Info("file-scanner-service service: Удаляем файл после сканирования",
		zap.String("scanID", scanMsg.ScanID),
		zap.String("file", scanMsg.FilePath),
	)
	// Теперь можно удалить файл:

	if err := os.Remove(filepath.Join(localDir, filepath.Base(scanMsg.FilePath))); err != nil {
		s.logger.Error("file-scanner-service service: Не удалось удалить файл",
			zap.String("scanID", scanMsg.ScanID),
			zap.String("file", scanMsg.FilePath),
			zap.Error(err),
		)
	} else {
		s.logger.Info("file-scanner-service service: Файл успешно удалён после сканирования",
			zap.String("scanID", scanMsg.ScanID),
			zap.String("file", scanMsg.FilePath),
		)
	}

	s.logger.Info("file-scanner-service service: Отправляем результат в топик сканирования",
		zap.String("scanID", scanMsg.ScanID),
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
			zap.String("scanID", scanMsg.ScanID),
			zap.Error(err),
		)
		filescannerservice.ErrorCounter.Inc()
		filescannerservice.ErrorSendingScanFilesResultDuration.Observe(time.Since(startTime).Seconds())
	}

	filescannerservice.SendingScanFilesResultsDuration.Observe(time.Since(startTime).Seconds())
	filescannerservice.SendedScanFilesResults.Inc()
	s.logger.Info("file-scanner-service service: Отправка результатов сканирования завершена",
		zap.String("scanID", scanMsg.ScanID),
	)

	countMessage := models.CountMessage{
		ScanID: scanMsg.ScanID,
		Number: 1,
	}

	s.logger.Info("file-scanner-service service: Отправляем количество завершенных файлов",
		zap.String("scanID", scanMsg.ScanID),
		zap.Any("countMessage", countMessage),
	)

	// Отправка 1, как количество завершенных файлов
	startTime = time.Now()

	if err := s.producer.SendMessage(s.config.KafkaCompletedFilesCountProducer.CompletedFilesCountTopic, countMessage); err != nil {
		s.logger.Error("file-scanner-service service: Ошибка при отправке количества отсканированных файлов в Kafka",
			zap.String("scanID", scanMsg.ScanID),
			zap.Error(err),
		)
		filescannerservice.ErrorSendingCompletedFilesDuration.Observe(time.Since(startTime).Seconds())
		filescannerservice.ErrorCounter.Inc()

	} else {
		s.logger.Info("file-scanner-service service: Отправка количества завершенных файлов завершена",
			zap.String("scanID", scanMsg.ScanID),
		)
		filescannerservice.SendingCompletedFilesDuration.Observe(time.Since(startTime).Seconds())
		filescannerservice.SendedCompletedFiles.Inc()
	}

	s.logger.Info("file-scanner-service service: Отправка количества завершенных файлов завершена",
		zap.String("scanID", scanMsg.ScanID),
	)

	return nil

}

func (s *fileScannerService) ReturnMessage(scanMsg *models.FileScanMessage) error {
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