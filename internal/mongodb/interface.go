package mongodb

import (
	"context"
	"ftp-scanner_try2/internal/models"
)

type ReportRepository interface {
	GetReportsByScanID(ctx context.Context, scanID string) ([]models.ScanReport, error)
}

type CounterRepository interface {
	GetCountersByScanID(ctx context.Context,  scanID string) (*models.CounterResponseGRPC, error)
}