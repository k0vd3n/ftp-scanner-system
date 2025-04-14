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
	if err := s.cleanupAndSendCounters(ctx, scanMsg, localDir); err != nil {
		s.logger.Error("file-scanner-service service: ProcessFile: Ошибка при удалении файла и отправке счетчиков", 
			zap.String("file", scanMsg.FilePath),
			zap.String("scan_id", scanMsg.ScanID))
		return err
	}
	s.logger.Info("file-scanner-service service: ProcessFile: файл успешно удален",
		zap.String("file", scanMsg.FilePath),
		zap.String("scan_id", scanMsg.ScanID))
	return nil

	/*
	   log.Printf("INFO: file-scanner-service service: kafkaMsg: %d, Сканируем файл: %s", counterOfAllMessages, scanMsg.FilePath)

	   // Создаём контекст с таймаутом (например, 30 секунд; при необходимости значение можно вынести в конфигурацию)
	   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	   defer cancel()

	   // filescannerservice.ReceivedMessages.Inc()
	   // Скачиваем файл
	   log.Printf("INFO: file-scanner-service service: kafkaMsg: %d, Скачиваем файл: %s", counterOfAllMessages, scanMsg.FilePath)
	   localDir := filepath.Join(s.config.KafkaScanResultProducer.FileScanDownloadPath, scanMsg.ScanID)
	   perm, err := strconv.ParseUint(s.config.KafkaScanResultProducer.Permision, 8, 32)

	   	if err != nil {
	   		return fmt.Errorf("ERROR: file-scanner-service service: kafkaMsg: %d, Ошибка при парсинге прав доступа: %v", counterOfAllMessages, err)
	   	}

	   fileMode := os.FileMode(perm)
	   log.Printf("INFO: file-scanner-service service: kafkaMsg: %d, Создание директории %s с правами %0o", counterOfAllMessages, localDir, perm)

	   	if err := os.MkdirAll(localDir, fileMode); err != nil {
	   		log.Printf("ERROR: file-scanner-service service: kafkaMsg: %d, Ошибка при создании директории %s: %v", counterOfAllMessages, localDir, err)
	   		return err
	   	}

	   resultChan := make(chan error, 1)

	   // Запускаем скачивание в отдельной горутине
	   log.Printf("INFO: file-scanner-service service: kafkaMsg: %d, Запускаем скачивание в отдельной горутине", counterOfAllMessages)

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

	   	log.Printf("WARNING: file-scanner-service: kafkaMsg: %d, Таймаут при скачивании файла %s", counterOfAllMessages, scanMsg.FilePath)
	   	log.Printf("INFO: file-scanner-service: kafkaMsg: %d, Возвращаем файл %s обратно в Kafka", counterOfAllMessages, scanMsg.FilePath)
	   	startTime := time.Now()
	   	if err := s.producer.SendMessage(s.config.KafkaConsumer.ConsumerTopic, models.FileScanMessage{
	   		ScanID:        scanMsg.ScanID,
	   		FilePath:      scanMsg.FilePath,
	   		ScanType:      scanMsg.ScanType,
	   		FTPConnection: scanMsg.FTPConnection,
	   	}); err != nil {
	   		log.Printf("ERROR: file-scanner-service: kafkaMsg: %d, Ошибка при возврате файла %s в топик Kafka: %v", counterOfAllMessages, scanMsg.FilePath, err)
	   		filescannerservice.ReturnErrorDuration.Observe(time.Since(startTime).Seconds())
	   		filescannerservice.ErrorCounter.Inc()
	   	} else {
	   		filescannerservice.ReturningFilesDuration.Observe(time.Since(startTime).Seconds())
	   		filescannerservice.ReturnedFiles.Inc()
	   		log.Printf("INFO: file-scanner-service: kafkaMsg: %d, Файл %s успешно возвращен в топик Kafka", counterOfAllMessages, scanMsg.FilePath)
	   	}
	   	return ctx.Err()

	   case err := <-resultChan:

	   		if err != nil {
	   			log.Printf("ERROR: file-scanner-service: kafkaMsg: %d, Ошибка при скачивании файла %s: %v", counterOfAllMessages, scanMsg.FilePath, err)
	   			return err
	   		}
	   	}

	   log.Printf("INFO: file-scanner-service: kafkaMsg: %d, Файл %s успешно скачан", counterOfAllMessages, scanMsg.FilePath)

	   log.Printf("INFO: file-scanner-service service: kafkaMsg: %d, Запускаем сканирование", counterOfAllMessages)
	   startScan := time.Now()
	   scanner, exists := s.scannerMap[scanMsg.ScanType]

	   	if !exists {
	   		log.Printf("ERROR: file-scanner-service service: kafkaMsg: %d, Тип сканирования не поддерживается: %s", counterOfAllMessages, scanMsg.ScanType)
	   		return nil
	   	}

	   result, err := scanner.Scan(filepath.Join(localDir, filepath.Base(scanMsg.FilePath)))

	   	if err != nil {
	   		log.Printf("ERROR: file-scanner-service service: kafkaMsg: %d, Ошибка при сканировании файла %s: %v", counterOfAllMessages, scanMsg.FilePath, err)
	   		return err
	   	}

	   log.Printf("INFO: file-scanner-service service: kafkaMsg: %d, Результат сканирования: %s", counterOfAllMessages, result)
	   filescannerservice.ScannedFiles.Inc()
	   scanTime := time.Since(startScan).Seconds()
	   filescannerservice.ScanDuration.Observe(scanTime)
	   log.Printf("INFO: file-scanner-service service: kafkaMsg: %d, Удаляем файл %s после сканирования", counterOfAllMessages, scanMsg.FilePath)
	   // Теперь можно удалить файл:

	   	if err := os.Remove(filepath.Join(localDir, filepath.Base(scanMsg.FilePath))); err != nil {
	   		log.Printf("ERROR: file-scanner-service service: kafkaMsg: %d, Не удалось удалить файл %s: %v", counterOfAllMessages, scanMsg.FilePath, err)
	   	} else {

	   		log.Printf("INFO: file-scanner-service service: kafkaMsg: %d, Файл %s успешно удалён после сканирования", counterOfAllMessages, scanMsg.FilePath)
	   	}

	   log.Printf("INFO: file-scanner-service service: kafkaMsg: %d, Отправляем результат в топик сканирования", counterOfAllMessages)
	   // Отправляем результат в Kafka
	   startTime := time.Now()

	   	if err := s.scanResultProducer.SendMessage(models.ScanResultMessage{
	   		ScanID:   scanMsg.ScanID,
	   		FilePath: scanMsg.FilePath,
	   		ScanType: scanMsg.ScanType,
	   		Result:   result,
	   	}); err != nil {

	   		log.Printf("ERROR: file-scanner-service service: kafkaMsg: %d, Ошибка при отправке результата сканирования в Kafka: %v", counterOfAllMessages, err)
	   		filescannerservice.ErrorCounter.Inc()
	   		filescannerservice.ErrorSendingScanFilesResultDuration.Observe(time.Since(startTime).Seconds())
	   		// return err
	   	}

	   filescannerservice.SendingScanFilesResultsDuration.Observe(time.Since(startTime).Seconds())
	   filescannerservice.SendedScanFilesResults.Inc()
	   log.Printf("INFO: file-scanner-service service: kafkaMsg: %d, Отправка результатов сканирования завершена", counterOfAllMessages)

	   // сообщение количества завершенных файлов

	   	countMessage := models.CountMessage{
	   		ScanID: scanMsg.ScanID,
	   		Number: 1,
	   	}

	   log.Printf("INFO: file-scanner-service service: kafkaMsg: %d, Отправляем количество завершенных файлов: %v", counterOfAllMessages, countMessage)
	   // Отправляем количество завершенных файлов
	   startTime = time.Now()

	   	if err := s.producer.SendMessage(s.config.KafkaCompletedFilesCountProducer.CompletedFilesCountTopic, countMessage); err != nil {
	   		log.Printf("ERROR: file-scanner-service service: kafkaMsg: %d, Ошибка при отправке количества отсканированных файлов в Kafka: %v", counterOfAllMessages, err)
	   		filescannerservice.ErrorSendingCompletedFilesDuration.Observe(time.Since(startTime).Seconds())
	   		filescannerservice.ErrorCounter.Inc()

	   		// return err
	   	} else {

	   		log.Printf("INFO: file-scanner-service service: kafkaMsg: %d, Отправка количества завершенных файлов завершена", counterOfAllMessages)
	   		filescannerservice.SendingCompletedFilesDuration.Observe(time.Since(startTime).Seconds())
	   		filescannerservice.SendedCompletedFiles.Inc()
	   	}

	   log.Printf("INFO: file-scanner-service service: kafkaMsg: %d, Отправка количества завершенных файлов завершена", counterOfAllMessages)

	   return nil
	*/
}

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

func (s *fileScannerService) cleanupAndSendCounters(ctx context.Context, scanMsg *models.FileScanMessage, localDir string) error {
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
