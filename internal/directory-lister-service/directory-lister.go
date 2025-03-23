package directorylisterservice

import (
	"ftp-scanner_try2/config"
	ftpclient "ftp-scanner_try2/internal/ftp"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"log"
)

type DirectoryListerService interface {
	ProcessDirectory(scanMsg *models.DirectoryScanMessage, ftpRepo ftpclient.FtpClientInterface) error
}

type directoryListerService struct {
	producer kafka.KafkaPoducerInterface
	config   config.DirectoryListerKafkaConfig
}

func NewDirectoryListerService(
	producer kafka.KafkaPoducerInterface,
	config config.DirectoryListerKafkaConfig,
) DirectoryListerService {
	return &directoryListerService{
		producer: producer,
		config:   config,
	}
}

func (s *directoryListerService) ProcessDirectory(scanMsg *models.DirectoryScanMessage, ftpRepo ftpclient.FtpClientInterface) error {

	log.Printf("Directory-lister-service: Сканируем директорию: %s", scanMsg.DirectoryPath)
	directories, files, err := ftpRepo.ListDirectory(scanMsg.DirectoryPath)
	if err != nil {
		log.Printf("Directory-lister-service: Ошибка при сканировании директории %s: %v", scanMsg.DirectoryPath, err)
		return err
	}

	log.Printf("Directory-lister-service: Найдено %d поддиректорий и %d файлов", len(directories), len(files))
	// Если поддиректорий не нашли, то не будем отправлять в топик листинга директорий 0
	if len(directories) != 0 {
		// Отправляем поддиректории в топик листинга директорий
		for _, dir := range directories {
			msg := models.DirectoryScanMessage{
				ScanID:        scanMsg.ScanID,
				DirectoryPath: dir,
				ScanTypes:     scanMsg.ScanTypes,
			}
			s.producer.SendMessage(s.config.DirectoriesToScanTopic, msg)
		}

		// Отправляем число найденных директорий в `scan-directories-count`
		s.producer.SendMessage(s.config.ScanDirectoriesCountTopic, models.CountMessage{
			ScanID: scanMsg.ScanID,
			Number: len(directories),
		})
	}

	ftpConnection := models.FTPConnection{
		Server:   scanMsg.FTPConnection.Server,
		Port:     scanMsg.FTPConnection.Port,
		Username: scanMsg.FTPConnection.Username,
		Password: scanMsg.FTPConnection.Password,
	}

	// Отправляем найденные файлы в `files-to-scan`
	for _, file := range files {
		for _, scanType := range scanMsg.ScanTypes {
			msg := models.FileScanMessage{
				ScanID:        scanMsg.ScanID,
				FilePath:      file,
				ScanType:      scanType,
				FTPConnection: ftpConnection,
			}
			s.producer.SendMessage(s.config.FilesToScanTopic, msg)
		}
	}

	// Отправляем число файлов в `scan-files-count`
	s.producer.SendMessage(s.config.ScanFilesCountTopic, models.CountMessage{
		ScanID: scanMsg.ScanID,
		Number: len(files) * len(scanMsg.ScanTypes),
	})

	// Отправляем завершение сканирования директории в `scan-completed-directories-count`
	s.producer.SendMessage(s.config.CompletedDirectoriesCountTopic, models.CountMessage{
		ScanID: scanMsg.ScanID,
		Number: 1,
	})

	return nil
}
