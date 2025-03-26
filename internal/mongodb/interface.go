package mongodb

import (
	"context"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/models"
)

type ReportRepository interface {
	GetReportsByScanID(ctx context.Context, scanID string) ([]models.ScanReport, error)
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
