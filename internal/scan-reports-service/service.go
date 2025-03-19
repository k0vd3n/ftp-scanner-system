package reportservice

import (
	"context"
	"ftp-scanner_try2/internal/models"
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

func GroupScanResults(reports []models.ScanReport) []models.ScanReport {
	scanMap := make(map[string]*models.ScanReport)

	for _, report := range reports {
		if _, exists := scanMap[report.ScanID]; !exists {
			scanMap[report.ScanID] = &models.ScanReport{
				ScanID:      report.ScanID,
				Directories: []models.Directory{},
			}
		}
		currentReport := scanMap[report.ScanID]
		currentReport.Directories = append(currentReport.Directories, report.Directories...)
	}

	var groupedResults []models.ScanReport
	for _, report := range scanMap {
		groupedResults = append(groupedResults, *report)
	}
	return groupedResults
}
