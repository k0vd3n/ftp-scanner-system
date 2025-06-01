package mongodb

import (
	"context"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/models"
)

type GenerateReportRepository interface {
	GetReportsByScanID(ctx context.Context, scanID string) ([]models.ScanReport, error)
	SaveResult(ctx context.Context, report models.ScanReport) error
}

type GetReportRepository interface {
	GetReport(ctx context.Context, scanID string) (*models.ScanReport, error)
}

type GetCounterRepository interface {
	GetCountersByScanID(ctx context.Context, scanID string, config config.StatusServiceMongo) (*models.StatusResponseGRPC, error)
}

type CounterReducerRepository interface {
	InsertReducedCounters(ctx context.Context, counts []models.CountMessage) error
}

type SaveReportRepository interface {
	InsertScanReports(ctx context.Context, reports []models.ScanReport) error
}

type MetricRepository interface {
	Save(ctx context.Context, instance string, payload []byte) error
	Load(ctx context.Context, instance string) ([]byte, error)
}
