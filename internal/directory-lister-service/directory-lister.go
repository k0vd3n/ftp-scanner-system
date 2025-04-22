package directorylisterservice

import (
	"context"
	"fmt"
	"ftp-scanner_try2/config"
	ftpclient "ftp-scanner_try2/internal/ftp"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"time"

	"go.uber.org/zap"
)

type DirectoryListerService interface {
	ProcessDirectory(ctx context.Context, scanMsg *models.DirectoryScanMessage, ftpRepo ftpclient.FtpClientInterface) error
}

type directoryListerService struct {
	producer kafka.KafkaPoducerInterface
	config   config.DirectoryListerKafkaConfig
	logger   *zap.Logger
}

func NewDirectoryListerService(
	producer kafka.KafkaPoducerInterface,
	config config.DirectoryListerKafkaConfig,
	logger *zap.Logger,
) DirectoryListerService {
	return &directoryListerService{
		producer: producer,
		config:   config,
		logger:   logger,
	}
}

func (s *directoryListerService) retryFTPListing(
	ctx context.Context,
	ftpRepo ftpclient.FtpClientInterface,
	directoryPath string,
	maxRetries int,
	timeoutPerAttempt time.Duration,
) ([]string, []string, error) {

	type result struct {
		dirs  []string
		files []string
		err   error
	}

	for attempt := range maxRetries {
		resultChan := make(chan result, 1)
		ftpCtx, ftpCancel := context.WithTimeout(ctx, timeoutPerAttempt)

		// Запускаем листинг директории в горутине
		s.logger.Info("Directory-lister-service: retryFTPListing - Запуск листинга в горутине, попытка",
			zap.Int("attempt", attempt+1),
			zap.String("directory", directoryPath))
		go func() {
			start := time.Now()
			dirs, files, err := ftpRepo.ListDirectory(directoryPath)
			FtpListDuration.Observe(time.Since(start).Seconds())
			resultChan <- result{dirs, files, err}
		}()

		select {
		case res := <-resultChan:
			ftpCancel()
			if res.err == nil {
				return res.dirs, res.files, nil
			}
			s.logger.Error("Directory-lister-service: retryFTPListing - Ошибка сканирования директории, попытка",
				zap.Int("attempt", attempt+1),
				zap.String("directory", directoryPath),
				zap.Error(res.err))
		case <-ftpCtx.Done():
			ftpCancel()
			s.logger.Warn("Directory-lister-service: retryFTPListing - Таймаут сканирования директории, попытка",
				zap.Int("attempt", attempt+1),
				zap.String("directory", directoryPath))
		}
	}

	return nil, nil, fmt.Errorf("directory-lister-service: retryFTPListing - Все %d попытки листинга директории завершились ошибкой", maxRetries)
}

func (s *directoryListerService) retryKafkaSend(
	ctx context.Context,
	topic string,
	msg any,
	maxRetries int,
	timeoutPerAttempt time.Duration,
	logString string,
) error {

	for attempt := range maxRetries {
		resultChan := make(chan error, 1)
		kafkaCtx, kafkaCancel := context.WithTimeout(ctx, timeoutPerAttempt)

		go func() {
			err := s.producer.SendMessage(topic, msg)
			resultChan <- err
		}()

		select {
		case err := <-resultChan:
			kafkaCancel()
			if err == nil {
				return nil
			}
			s.logger.Error("Directory-lister-service: retryKafkaSend - Ошибка отправки в Kafka "+logString,
				zap.Int("attempt", attempt+1),
				zap.String("topic", topic),
				zap.Error(err))
		case <-kafkaCtx.Done():
			kafkaCancel()
			s.logger.Warn("Directory-lister-service: retryKafkaSend - Таймаут отправки в Kafka "+logString,
				zap.Int("attempt", attempt+1),
				zap.String("topic", topic))
		}
	}

	return fmt.Errorf("Directory-lister-service: retryKafkaSend - Ошибка отправки в Kafka топик %s после %d попыток "+logString, topic, maxRetries)
}

func (s *directoryListerService) ProcessDirectory(ctx context.Context, scanMsg *models.DirectoryScanMessage, ftpRepo ftpclient.FtpClientInterface) error {

	s.logger.Info("Directory-lister-service: ProcessDirectory - Сканируем директорию",
		zap.String("directory", scanMsg.DirectoryPath),
		zap.String("scan_id", scanMsg.ScanID),
		zap.String("method", "ProcessDirectory"))

	dirs, files, err := s.retryFTPListing(
		ctx,
		ftpRepo,
		scanMsg.DirectoryPath,
		s.config.MaxRetries,
		time.Duration(s.config.TimeoutSeconds)*time.Second,
	)
	if err != nil {
		msg := models.DirectoryScanMessage{
			ScanID:        scanMsg.ScanID,
			DirectoryPath: scanMsg.DirectoryPath,
			ScanTypes:     scanMsg.ScanTypes,
			FTPConnection: scanMsg.FTPConnection,
		}

		logString := fmt.Sprintf("возврата директории %s в топик после неудавшегося листинга", scanMsg.DirectoryPath)
		if err := s.retryKafkaSend(
			ctx,
			s.config.DirectoriesToScanTopic,
			msg,
			s.config.MaxRetries,
			time.Duration(s.config.TimeoutSeconds)*time.Second,
			logString,
		); err != nil {
			KafkaSendErrors.Inc()
			s.logger.Error("Directory-lister-service: ProcessDirectory - Ошибка возврата директории в топик",
				zap.String("directory", scanMsg.DirectoryPath),
				zap.Error(err))
		} else {
			KafkaMessagesReturnedDirs.Inc()
		}
		return err
	}
	DirectoriesProcessed.Inc()
	if len(dirs) != 0 {
		for _, dir := range dirs {
			msg := models.DirectoryScanMessage{
				ScanID:        scanMsg.ScanID,
				DirectoryPath: dir,
				ScanTypes:     scanMsg.ScanTypes,
				FTPConnection: scanMsg.FTPConnection,
			}
			logString := fmt.Sprintf("найденной поддиректории %s в топик на листинг", dir)
			if err := s.retryKafkaSend(
				ctx,
				s.config.DirectoriesToScanTopic,
				msg,
				s.config.MaxRetries,
				time.Duration(s.config.TimeoutSeconds)*time.Second,
				logString,
			); err != nil {
				KafkaSendErrors.Inc()
				s.logger.Error("Directory-lister-service: ProcessDirectory - Ошибка отправки в Kafka поддиректории в топик",
					zap.String("directory", dir),
					zap.String("topic", s.config.DirectoriesToScanTopic),
					zap.String("scan_id", scanMsg.ScanID),
					zap.Error(err))
			} else {
				s.logger.Info("Directory-lister-service: ProcessDirectory - Поддиректория на листинг отправлена в Kafka",
					zap.String("directory", dir),
					zap.String("topic", s.config.DirectoriesToScanTopic),
					zap.String("scan_id", scanMsg.ScanID))
				KafkaMessagesSentDirectories.Inc()
			}

		}
		msg := models.CountMessage{
			ScanID: scanMsg.ScanID,
			Number: len(dirs),
		}
		logString := fmt.Sprintf("количества найденных директорий в процессе листинга директории %s", scanMsg.DirectoryPath)

		if err := s.retryKafkaSend(
			ctx,
			s.config.ScanDirectoriesCountTopic,
			msg,
			s.config.MaxRetries,
			time.Duration(s.config.TimeoutSeconds)*time.Second,
			logString,
		); err != nil {
			KafkaSendErrors.Inc()
			s.logger.Error("Directory-lister-service: ProcessDirectory - Ошибка отправки в Kafka топик количества найденных директорий",
				zap.String("directory", scanMsg.DirectoryPath),
				zap.String("topic", s.config.DirectoriesToScanTopic),
				zap.String("scan_id", scanMsg.ScanID),
				zap.Error(err))
		} else {
			s.logger.Info("Directory-lister-service: ProcessDirectory - Количество найденных директорий отправлено в Kafka",
				zap.String("directory", scanMsg.DirectoryPath),
				zap.String("topic", s.config.DirectoriesToScanTopic),
				zap.String("scan_id", scanMsg.ScanID))
			KafkaMessagesSentScanDirsCount.Add(float64(len(dirs)))
		}
	}

	if len(files) != 0 {
		FilesFound.Add(float64(len(files)))
		for _, file := range files {
			for _, scanType := range scanMsg.ScanTypes {
				msg := models.FileScanMessage{
					ScanID:        scanMsg.ScanID,
					FilePath:      file,
					ScanType:      scanType,
					FTPConnection: scanMsg.FTPConnection,
				}
				logString := fmt.Sprintf("найденного файла %s в процессе листинга директории", file)
				if err := s.retryKafkaSend(
					ctx,
					s.config.FilesToScanTopic,
					msg,
					s.config.MaxRetries,
					time.Duration(s.config.TimeoutSeconds)*time.Second,
					logString,
				); err != nil {
					s.logger.Error("Directory-lister-service: ProcessDirectory - Ошибка отправки файла в Kafka",
						zap.String("file", file),
						zap.Error(err))
					KafkaSendErrors.Inc()
				} else {
					s.logger.Info("Directory-lister-service: ProcessDirectory - Файл отправлен в Kafka",
						zap.String("file", file))
					KafkaMessagesSentFiles.Inc()
				}
			}
		}
		logString := fmt.Sprintf("количества найденных файлов в процессе листинга директории %s", scanMsg.DirectoryPath)
		msg := models.CountMessage{
			ScanID: scanMsg.ScanID,
			Number: len(files),
		}
		if err := s.retryKafkaSend(
			ctx,
			s.config.ScanFilesCountTopic,
			msg,
			s.config.MaxRetries,
			time.Duration(s.config.TimeoutSeconds)*time.Second,
			logString,
		); err != nil {
			KafkaSendErrors.Inc()
			s.logger.Error("Directory-lister-service: ProcessDirectory - Ошибка отправки в Kafka топик количества найденных файлов",
				zap.String("directory", scanMsg.DirectoryPath),
				zap.String("topic", s.config.DirectoriesToScanTopic),
				zap.String("scan_id", scanMsg.ScanID),
				zap.Error(err))
		} else {
			s.logger.Info("Directory-lister-service: ProcessDirectory - Количество найденных файлов отправлено в Kafka",
				zap.String("directory", scanMsg.DirectoryPath),
				zap.String("topic", s.config.DirectoriesToScanTopic),
				zap.String("scan_id", scanMsg.ScanID))
			KafkaMessagesSentScanFilesCount.Add(float64(len(files)))
		}
	}

	logString := fmt.Sprintf("завершения обработки директории %s", scanMsg.DirectoryPath)
	msg := models.CountMessage{
		ScanID: scanMsg.ScanID,
		Number: 1,
	}
	if err := s.retryKafkaSend(
		ctx,
		s.config.CompletedDirectoriesCountTopic,
		msg,
		s.config.MaxRetries,
		time.Duration(s.config.TimeoutSeconds)*time.Second,
		logString,
	); err != nil {
		KafkaSendErrors.Inc()
		s.logger.Error("Directory-lister-service: ProcessDirectory - Ошибка отправки в Kafka завершения обработки директории",
			zap.String("directory", scanMsg.DirectoryPath),
			zap.String("topic", s.config.DirectoriesToScanTopic),
			zap.String("scan_id", scanMsg.ScanID),
			zap.Error(err))
	} else {
		s.logger.Info("Directory-lister-service: ProcessDirectory - Обработка директории завершена",
			zap.String("directory", scanMsg.DirectoryPath),
			zap.String("topic", s.config.DirectoriesToScanTopic),
			zap.String("scan_id", scanMsg.ScanID))
		KafkaMessagesSentCompletedDirsCount.Inc()
	}

	return nil
}
