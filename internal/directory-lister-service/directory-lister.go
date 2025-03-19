package directorylisterservice

import (
	"fmt"
	"ftp-scanner_try2/config"
	ftpclient "ftp-scanner_try2/internal/ftp"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"log"

	"github.com/joho/godotenv"
)

type DirectoryListerService interface {
	ProcessDirectory(scanMsg *models.DirectoryScanMessage, ftpRepo ftpclient.FtpClientInterface) error
}

type directoryListerService struct {
	// ftpRepo  ftpclient.FtpClientInterface
	producer kafka.KafkaPoducerInterface
	config   config.DirectoryListerConfig
}

func NewDirectoryListerService( /*ftpRepo ftpclient.FtpClientInterface,*/ producer kafka.KafkaPoducerInterface, config config.DirectoryListerConfig) DirectoryListerService {
	return &directoryListerService{
		// ftpRepo:  ftpRepo,
		producer: producer,
		config:   config,
	}
}

func (s *directoryListerService) ProcessDirectory(scanMsg *models.DirectoryScanMessage, ftpRepo ftpclient.FtpClientInterface) error {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Подключение к FTP серверу
	ftpClient, err := ftpclient.NewFTPClient(fmt.Sprintf("%s:%d", scanMsg.FTPConnection.Server, scanMsg.FTPConnection.Port), scanMsg.FTPConnection.Username, scanMsg.FTPConnection.Password)
	if err != nil {
		log.Fatal("Ошибка подключения к FTP:", err)
	}
	defer ftpClient.Close()

	log.Printf("Сканируем директорию: %s", scanMsg.DirectoryPath)
	directories, files, err := ftpRepo.ListDirectory(scanMsg.DirectoryPath)
	if err != nil {
		log.Printf("Ошибка при сканировании директории %s: %v", scanMsg.DirectoryPath, err)
		return err
	}

	log.Printf("Найдено %d поддиректорий и %d файлов", len(directories), len(files))

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
