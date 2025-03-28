package service

import (
	"fmt"
	"ftp-scanner_try2/config"
	filescannerservice "ftp-scanner_try2/internal/file-scanner-service"
	"ftp-scanner_try2/internal/file-scanner-service/scanner"
	ftpclient "ftp-scanner_try2/internal/ftp"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type FileScannerService interface {
	ProcessFile(scanMsg *models.FileScanMessage, ftpClient ftpclient.FtpClientInterface) error
}

type fileScannerService struct {
	scanResultProducer kafka.KafkaScanResultPoducerInterface
	counterProducer    kafka.KafkaPoducerInterface
	config             config.FileScannerConfig
	scannerMap         map[string]scanner.FileScanner
}

func NewFileScannerService(
	scanResultProducer kafka.KafkaScanResultPoducerInterface,
	counterProducer kafka.KafkaPoducerInterface,
	config config.FileScannerConfig,
	scannerMap map[string]scanner.FileScanner,
) FileScannerService {
	return &fileScannerService{
		scanResultProducer: scanResultProducer,
		counterProducer:    counterProducer,
		config:             config,
		scannerMap:         scannerMap,
	}
}

func (s *fileScannerService) ProcessFile(scanMsg *models.FileScanMessage, ftpClient ftpclient.FtpClientInterface) error {
	log.Printf("file-scanner-service service: Сканируем файл: %s", scanMsg.FilePath)
	filescannerservice.ReceivedMessages.Inc()
	// Скачиваем файл
	localDir := filepath.Join(s.config.KafkaScanResultProducer.FileScanDownloadPath, scanMsg.ScanID)
	perm, err := strconv.ParseUint(s.config.KafkaScanResultProducer.Permision, 8, 32)
	if err != nil {
		return fmt.Errorf("file-scanner-service service: Ошибка при парсинге прав доступа: %v", err)
	}
	fileMode := os.FileMode(perm)
	log.Printf("file-scanner-service service: Создание директории %s с правами %0o", localDir, perm)
	if err := os.MkdirAll(localDir, fileMode); err != nil {
		log.Printf("file-scanner-service service: Ошибка при создании директории %s: %v", localDir, err)
		return err
	}

	startDownload := time.Now()
	if err := ftpClient.DownloadFile(scanMsg.FilePath, localDir); err != nil {
		log.Printf("file-scanner-service service: Ошибка при скачивании файла %s: %v", scanMsg.FilePath, err)
		return err
	}

	// Сканируем файл
	downloadTime := time.Since(startDownload).Seconds()
	filescannerservice.DownloadDuration.Observe(downloadTime)

	startScan := time.Now()
	scanner, exists := s.scannerMap[scanMsg.ScanType]
	if !exists {
		log.Printf("file-scanner-service service: Тип сканирования не поддерживается: %s", scanMsg.ScanType)
		return nil
	}

	result, err := scanner.Scan(filepath.Join(localDir, filepath.Base(scanMsg.FilePath)))
	if err != nil {
		log.Printf("file-scanner-service service: Ошибка при сканировании файла %s: %v", scanMsg.FilePath, err)
		return err
	}
	scanTime := time.Since(startScan).Seconds()
	filescannerservice.ScanDuration.Observe(scanTime)
	log.Printf("file-scanner-service service: Результат сканирования: %s", result)
	log.Printf("file-scanner-service service: Отправляем результат в топик сканирования")
	// Отправляем результат в Kafka
	if err := s.scanResultProducer.SendMessage(models.ScanResultMessage{
		ScanID:   scanMsg.ScanID,
		FilePath: scanMsg.FilePath,
		ScanType: scanMsg.ScanType,
		Result:   result,
	}); err != nil {
		log.Printf("file-scanner-service service: Ошибка при отправке результата сканирования в Kafka: %v", err)
		filescannerservice.ErrorCounter.Inc()
		// return err
	}
	log.Printf("file-scanner-service service: Отправка результатов сканирования завершена")

	// сообщение количества завершенных файлов
	countMessage := models.CountMessage{
		ScanID: scanMsg.ScanID,
		Number: 1,
	}
	log.Printf("file-scanner-service service: Отправляем количество завершенных файлов: %v", countMessage)
	// Отправляем количество завершенных файлов
	if err := s.counterProducer.SendMessage(s.config.KafkaCompletedFilesCountProducer.CompletedFilesCountTopic, countMessage); err != nil {
		log.Printf("file-scanner-service service: Ошибка при отправке количества отсканированных файлов в Kafka: %v", err)
		filescannerservice.ErrorCounter.Inc()
		// return err
	} else {
		filescannerservice.CompletedFilesSent.Inc()
	}
	log.Printf("file-scanner-service service: Отправка количества завершенных файлов завершена")

	return nil
}
