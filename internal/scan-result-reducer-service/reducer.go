package scanresultreducerservice

import (
	"context"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"ftp-scanner_try2/internal/mongodb"
	"strings"
	"time"

	"go.uber.org/zap"
)

type ReducerService interface {
	ReduceScanResults(messages []models.ScanResultMessage) []models.ScanReport
	Start(ctx context.Context)
}

type reducerService struct {
	repo     mongodb.SaveReportRepository
	cfg      config.UnifiedConfig
	consumer kafka.KafkaScanResultReducerConsumerInterface
	logger   *zap.Logger
}

func NewReducerService(
	repo mongodb.SaveReportRepository,
	cfg config.UnifiedConfig,
	consumer kafka.KafkaScanResultReducerConsumerInterface,
	logger *zap.Logger,
) ReducerService {
	return &reducerService{
		repo:     repo,
		cfg:      cfg,
		consumer: consumer,
		logger:   logger,
	}
}

func (r *reducerService) ReduceScanResults(messages []models.ScanResultMessage) []models.ScanReport {
	r.logger.Info("scan-result-reducer-service reduce-scan-results: Начало редьюса результатов сканирования...")
	start := time.Now()
	scanMap := make(map[string]*models.ScanReport)

	r.logger.Info("scan-result-reducer-service reduce-scan-results: Начало подсчета общего числа полученных отчетов для редьюса...")
	// Счётчики для директорий и файлов
	var totalDirs, totalFiles int

	r.logger.Info("scan-result-reducer-service reduce-scan-results: Начало редьюса результатов сканирования...")
	for _, msg := range messages {
		// Проверяем, существует ли отчет для данного scan_id
		if _, exists := scanMap[msg.ScanID]; !exists {
			scanMap[msg.ScanID] = &models.ScanReport{
				ScanID:      msg.ScanID,
				Directories: []models.Directory{},
			}
		}

		// Получаем указатель на текущий отчет
		currentReport := scanMap[msg.ScanID]
		// Разбиваем путь на директории и имя файла
		fullPath, dirs, _ := splitPath(msg.FilePath)

		// Проверяем, есть ли уже такая директория
		var currentDir *models.Directory
		for i := range currentReport.Directories {
			if currentReport.Directories[i].Directory == fullPath {
				currentDir = &currentReport.Directories[i]
				break
			}
		}

		// Если директории нет, создаем ее с абсолютным путем и увеличиваем счетчик директорий
		if currentDir == nil {
			newDir := models.Directory{
				Directory:    fullPath,
				Subdirectory: []models.Directory{},
				Files:        []models.File{},
			}
			currentReport.Directories = append(currentReport.Directories, newDir)
			currentDir = &currentReport.Directories[len(currentReport.Directories)-1]
			totalDirs++ // учет новой корневой директории
		}

		// Обрабатываем поддиректории, создавая полные пути,
		// начинаем с корневого пути fullPath, а не с пустой строки
		basePath := fullPath
		for _, dir := range dirs {
			basePath = basePath + "/" + dir
			found := false
			for i := range currentDir.Subdirectory {
				if currentDir.Subdirectory[i].Directory == basePath {
					currentDir = &currentDir.Subdirectory[i]
					found = true
					break
				}
			}
			if !found {
				newDir := models.Directory{
					Directory:    basePath,
					Subdirectory: []models.Directory{},
					Files:        []models.File{},
				}
				currentDir.Subdirectory = append(currentDir.Subdirectory, newDir)
				currentDir = &currentDir.Subdirectory[len(currentDir.Subdirectory)-1]
				totalDirs++ // учет новой директории
			}
		}

		// Добавляем файл в текущую директорию
		fileExists := false
		for i := range currentDir.Files {
			if currentDir.Files[i].Path == msg.FilePath {
				currentDir.Files[i].ScanResults = append(currentDir.Files[i].ScanResults, models.ScanResult{
					Type:   msg.ScanType,
					Result: msg.Result,
				})
				fileExists = true
				totalFiles++
				break
			}
		}
		if !fileExists {
			newFile := models.File{
				Path: msg.FilePath,
				ScanResults: []models.ScanResult{
					{
						Type:   msg.ScanType,
						Result: msg.Result,
					},
				},
			}
			currentDir.Files = append(currentDir.Files, newFile)
			totalFiles++ // учет нового файла
		}
	}

	ReceivedDirectories.Add(float64(totalDirs))
	ReceivedFiles.Add(float64(totalFiles))

	// Преобразуем map в слайс для JSON-ответа
	var groupedResults []models.ScanReport
	for _, report := range scanMap {
		groupedResults = append(groupedResults, *report)
	}
	duration := time.Since(start)
	ProcessingDuration.Observe(duration.Seconds())
	ReportsTotal.Add(float64(len(groupedResults)))
	r.logger.Info("scan-result-reducer-service reduce-scan-results: Редьюс завершен",
		zap.Int("reports", len(groupedResults)),
		zap.Duration("duration", duration))
	return groupedResults
}

func (r *reducerService) Start(ctx context.Context) {
	for {
		messages, totalMessages, err := r.consumer.ReadMessages(
			ctx,
			r.cfg.ScanResultReducer.Kafka.BatchSize,
			time.Duration(r.cfg.ScanResultReducer.Kafka.Duration)*time.Second,
		)
		if err != nil {
			r.logger.Error("scan-result-reducer-service main: Ошибка чтения сообщений из Kafka", zap.Error(err))
			continue
		}
		r.logger.Info("scan-result-reducer-service main: Получено сообщений из Kafka", zap.Int("count", len(messages)))

		if totalMessages > 0 {
			ReceivedMessages.Add(float64(totalMessages))
			r.logger.Info("scan-result-reducer-service main: Редьюс сообщений...", zap.Int("count", totalMessages))
			reducedResults := r.ReduceScanResults(messages)
			ReceivedMessagesKafkaHandler.Add(float64(totalMessages))
			r.logger.Info("scan-result-reducer-service main: Редьюс сообщений завершен", zap.Int("count", len(reducedResults)))
			if err := r.repo.InsertScanReports(ctx, reducedResults); err != nil {
				r.logger.Error("scan-result-reducer-service main: Ошибка сохранения данных в MongoDB", zap.Error(err))
				ErrorsTotal.Inc()
			}

			r.logger.Info("scan-result-reducer-service main: Сообщения успешно сохранены в MongoDB", zap.Int("count", len(reducedResults)))
		}
	}
}

// Разбиваем путь на абсолютный путь, список директорий и имя файла
func splitPath(filePath string) (string, []string, string) {
	parts := strings.Split(filePath, "/")
	if len(parts) < 2 {
		return filePath, nil, ""
	}
	root := "/" + parts[1]          // Корневой каталог
	fileName := parts[len(parts)-1] // Имя файла

	// Если есть поддиректории, исключаем дублирование корневой директории
	if len(parts) > 2 {
		// parts[1] соответствует корневой директории, поэтому начинаем с parts[2]
		return root, parts[2 : len(parts)-1], fileName
	}
	return root, nil, fileName
}
