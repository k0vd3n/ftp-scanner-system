package reportservice

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

// Интерфейс для хранилища отчетов
type ReportStorage interface {
	SaveReport(scanID string, data interface{}) (string, error)
}

// Локальное хранилище (файловая система)
type FileReportStorage struct {
	basePath string
	logger   *zap.Logger
}

func NewFileReportStorage(basePath string, logger *zap.Logger) *FileReportStorage {
	return &FileReportStorage{basePath: basePath, logger: logger}
}

func (s *FileReportStorage) SaveReport(scanID string, data interface{}) (string, error) {
	s.logger.Info("scan-reports-service storage: SaveReport: Сохранение отчета для scan_id=" + scanID)
	filePath := fmt.Sprintf("%s/%s_timestamp_%s.json",
		s.basePath,
		scanID,
		fmt.Sprintf(
			"%04d:%02d:%02d:%02d:%02d:%02d",
			time.Now().In(time.FixedZone("MSK", 3*60*60)).Year(),
			time.Now().In(time.FixedZone("MSK", 3*60*60)).Month(),
			time.Now().In(time.FixedZone("MSK", 3*60*60)).Day(),
			time.Now().In(time.FixedZone("MSK", 3*60*60)).Hour(),
			time.Now().In(time.FixedZone("MSK", 3*60*60)).Minute(),
			time.Now().In(time.FixedZone("MSK", 3*60*60)).Second(),
		),
	)

	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return "", err
	}

	s.logger.Info("scan-reports-service storage: SaveReport: Сохранение отчета для scan_id=" + scanID + " завершено")
	return filePath, nil
}
