package directorylisterservice

import (
	ftpclient "ftp-scanner_try2/internal/ftp"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"log"
)

type DirectoryListerService interface {
	ProcessDirectory(scanMsg *models.DirectoryScanMessage) error
}

type directoryListerService struct {
	ftpRepo  ftpclient.FtpClientInterface
	producer kafka.KafkaPoducerInterface
}

func NewDirectoryListerService(ftpRepo ftpclient.FtpClientInterface, producer kafka.KafkaPoducerInterface) DirectoryListerService {
	return &directoryListerService{
		ftpRepo:  ftpRepo,
		producer: producer,
	}
}

func (s *directoryListerService) ProcessDirectory(scanMsg *models.DirectoryScanMessage) error {
	log.Printf("Сканируем директорию: %s", scanMsg.DirectoryPath)
	directories, files, err := s.ftpRepo.ListDirectory(scanMsg.DirectoryPath)
	if err != nil {
		log.Printf("Ошибка при сканировании директории %s: %v", scanMsg.DirectoryPath, err)
		return err
	}

	log.Printf("Найдено %d поддиректорий и %d файлов", len(directories), len(files))

	// Отправляем поддиректории в `directories-to-scan`
	for _, dir := range directories {
		msg := models.DirectoryScanMessage{
			ScanID:        scanMsg.ScanID,
			DirectoryPath: dir,
			ScanTypes:     scanMsg.ScanTypes,
		}
		s.producer.SendMessage("directories-to-scan", msg)
	}

	// Отправляем число найденных директорий в `scan-directories-count`
	s.producer.SendMessage("scan-directories-count", models.CountMessage{
		ScanID: scanMsg.ScanID,
		Number: len(directories),
	})

	// Отправляем найденные файлы в `files-to-scan`
	for _, file := range files {
		for _, scanType := range scanMsg.ScanTypes {
			msg := models.FileScanMessage{
				ScanID:   scanMsg.ScanID,
				FilePath: file,
				ScanType: scanType,
			}
			s.producer.SendMessage("files-to-scan", msg)
		}
	}

	// Отправляем число файлов в `scan-files-count`
	s.producer.SendMessage("scan-files-count", models.CountMessage{
		ScanID: scanMsg.ScanID,
		Number: len(files) * len(scanMsg.ScanTypes),
	})

	// Отправляем завершение сканирования директории в `scan-completed-directories-count`
	s.producer.SendMessage("scan-completed-directories-count", models.CountMessage{
		ScanID: scanMsg.ScanID,
		Number: 1,
	})

	return nil
}