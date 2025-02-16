package reportservice

import "ftp-scanner_try2/internal/models"

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
