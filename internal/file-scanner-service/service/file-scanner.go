package service

import (
	"ftp-scanner_try2/config"
	ftpclient "ftp-scanner_try2/internal/ftp"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"log"
	"os"
	"path/filepath"
)

type FileScannerService interface {
	ProcessFile(scanMsg *models.FileScanMessage, ftpClient ftpclient.FtpClientInterface) error
}

type fileScannerService struct {
	// ftpClient  ftpclient.FtpClientInterface
	producer kafka.KafkaPoducerInterface
	// scannerMap map[string]scanner.FileScanner
	config config.FilesScannerConfig
}

func NewFileScannerService( /*ftpClient ftpclient.FtpClientInterface,*/ producer kafka.KafkaPoducerInterface, config config.FilesScannerConfig) FileScannerService {
	// scannerMap := make(map[string]scanner.FileScanner)
	// scannerMap["zero_bytes"] = scanner.NewZeroBytesScanner() // Регистрируем сканер для нулевых байтов

	return &fileScannerService{
		// ftpClient:  ftpClient,
		producer: producer,
		// scannerMap: scannerMap,
		config: config,
	}
}

func (s *fileScannerService) ProcessFile(scanMsg *models.FileScanMessage, ftpClient ftpclient.FtpClientInterface) error {
	log.Printf("Сканируем файл: %s", scanMsg.FilePath)

	// Скачиваем файл
	localDir := filepath.Join(s.config.PathForDownloadedFiles, scanMsg.ScanID)
	if err := os.MkdirAll(localDir, s.config.Permision); err != nil {
		return err
	}

	if err := ftpClient.DownloadFile(scanMsg.FilePath, localDir); err != nil {
		log.Printf("Ошибка при скачивании файла %s: %v", scanMsg.FilePath, err)
		return err
	}

	// Сканируем файл
	scanner, exists := s.config.ScannerTypesMap[scanMsg.ScanType]
	if !exists {
		log.Printf("Тип сканирования не поддерживается: %s", scanMsg.ScanType)
		return nil
	}

	result, err := scanner.Scan(filepath.Join(localDir, filepath.Base(scanMsg.FilePath)))
	if err != nil {
		log.Printf("Ошибка при сканировании файла %s: %v", scanMsg.FilePath, err)
		return err
	}

	// Отправляем результат в Kafka
	s.producer.SendMessage(s.config.ScanResultsTopic, models.ScanResultMessage{
		ScanID:   scanMsg.ScanID,
		FilePath: scanMsg.FilePath,
		ScanType: scanMsg.ScanType,
		Result:   result,
	})

	// Отправляем счетчик завершенных файлов
	s.producer.SendMessage(s.config.CompletedFilesCountTopic, models.CountMessage{
		ScanID: scanMsg.ScanID,
		Number: 1,
	})

	return nil
}
