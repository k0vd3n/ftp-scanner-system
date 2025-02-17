
package directorylisterservice
/*
import (
	ftpclient "ftp-scanner_try2/internal/ftp"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type DirectoryListerService struct {
	ftpClient ftpclient.FtpClientInterface
	producer  kafka.KafkaPoducerInterface
}

func NewDirectoryListerService(ftpClient ftpclient.FtpClientInterface, producer kafka.KafkaPoducerInterface) *DirectoryListerService {
	return &DirectoryListerService{
		ftpClient: ftpClient,
		producer:  producer,
	}
}

func (s *DirectoryListerService) ProcessDirectory(scanMsg *models.DirectoryScanMessage) error {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	directoriesToScan := os.Getenv("DIRECTORIES_TO_SCAN")
	scanDirectoriesCount := os.Getenv("SCAN_DIRECTORIES_COUNT")
	filesToScan := os.Getenv("FILES_TO_SCAN")
	scanFilesCount := os.Getenv("SCAN_FILES_COUNT")
	completedDirectoriesCount := os.Getenv("COMPLETED_DIRECTORIES_COUNT")

	log.Printf("Сканируем директорию: %s", scanMsg.DirectoryPath)
	directories, files, err := s.ftpClient.ListDirectory(scanMsg.DirectoryPath)
	if err != nil {
		log.Printf("Ошибка при сканировании директории %s: %v", scanMsg.DirectoryPath, err)
		return err
	}

	log.Printf("Найдено %d поддиректорий и %d файлов", len(directories), len(files))

	// 1 Отправляем поддиректории в `directories-to-scan`
	for _, dir := range directories {
		msg := models.DirectoryScanMessage{
			ScanID:        scanMsg.ScanID,
			DirectoryPath: dir,
			ScanTypes:     scanMsg.ScanTypes,
		}
		s.producer.SendMessage(directoriesToScan, msg)
	}

	// 2️ Отправляем число найденных директорий в `scan-directories-count`
	s.producer.SendMessage(scanDirectoriesCount, models.CountMessage{
		ScanID: scanMsg.ScanID,
		Number: len(directories),
	})
	log.Printf("Отправлено количество директорий в `%s`: %d", scanDirectoriesCount, len(directories))

	// 3️ Отправляем найденные файлы в `files-to-scan`
	for _, file := range files {
		for _, scanType := range scanMsg.ScanTypes {
			msg := models.FileScanMessage{
				ScanID:   scanMsg.ScanID,
				FilePath: file,
				ScanType: scanType,
			}
			s.producer.SendMessage(filesToScan, msg)
			log.Printf("Файл %s отправлен в `%s`", filesToScan, file)
		}
	}

	// 4️ Отправляем число файлов в `scan-files-count`
	s.producer.SendMessage(scanFilesCount, models.CountMessage{
		ScanID: scanMsg.ScanID,
		Number: len(files) * len(scanMsg.ScanTypes),
	})
	log.Printf("Отправлено количество файлов в `%s`: %d", scanFilesCount, len(files)*len(scanMsg.ScanTypes))

	// 5️ Отправляем завершение сканирования директории в `scan-completed-directories-count`
	s.producer.SendMessage(completedDirectoriesCount, models.CountMessage{
		ScanID: scanMsg.ScanID,
		Number: 1,
	})

	return nil
}

*/