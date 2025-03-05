package reportservice

import (
	"context"
	"ftp-scanner_try2/internal/mongodb"
)

type ReportServiceInterface interface {
	GenerateReport(ctx context.Context, scanID string) (string, error)
}

type ReportService struct {
	repo    mongodb.ReportRepository
	storage ReportStorage
}

func NewReportService(repo mongodb.ReportRepository, storage ReportStorage) *ReportService {
	return &ReportService{repo: repo, storage: storage}
}

func (s *ReportService) GenerateReport(ctx context.Context, scanID string) (string, error) {
	reports, err := s.repo.GetReportsByScanID(ctx, scanID)
	if err != nil {
		return "", err
	}

	groupedData := GroupScanResults(reports)

	return s.storage.SaveReport(scanID, groupedData)
}
