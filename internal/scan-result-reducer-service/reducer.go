package scanresultreducerservice

import (
	"ftp-scanner_try2/internal/models"
	"strings"
)

type ReducerService interface {
	ReduceScanResults(messages []models.ScanResultMessage) []models.ScanReport
}

type reducerService struct {
}

func NewReducerService() ReducerService {
	return &reducerService{}
}

func (r *reducerService) ReduceScanResults(messages []models.ScanResultMessage) []models.ScanReport {
	scanMap := make(map[string]*models.ScanReport)

	for _, msg := range messages {
		if _, exists := scanMap[msg.ScanID]; !exists {
			scanMap[msg.ScanID] = &models.ScanReport{
				ScanID:      msg.ScanID,
				Directories: []models.Directory{},
			}
		}
		currentReport := scanMap[msg.ScanID]

		dir, subDirs, _ := splitPath(msg.FilePath)
		var currentDir *models.Directory

		for i := range currentReport.Directories {
			if currentReport.Directories[i].Directory == dir {
				currentDir = &currentReport.Directories[i]
				break
			}
		}

		if currentDir == nil {
			newDir := models.Directory{
				Directory:    dir,
				Subdirectory: []models.Directory{},
				Files:        []models.File{},
			}
			currentReport.Directories = append(currentReport.Directories, newDir)
			currentDir = &currentReport.Directories[len(currentReport.Directories)-1]
		}

		basePath := ""
		for _, subDir := range subDirs {
			basePath += basePath + "/" + subDir
			found := false
			for i := range currentDir.Subdirectory {
				if currentDir.Subdirectory[i].Directory == basePath {
					currentDir = &currentDir.Subdirectory[i]
					found = true
					break
				}
			}
			if !found {
				newSubDir := models.Directory{
					Directory: basePath,
				}
				currentDir.Subdirectory = append(currentDir.Subdirectory, newSubDir)
				currentDir = &currentDir.Subdirectory[len(currentDir.Subdirectory)-1]
			}
		}

		fileExists := false
		for i := range currentDir.Files {
			if currentDir.Files[i].Path == msg.FilePath {
				currentDir.Files[i].ScanResults = append(currentDir.Files[i].ScanResults, models.ScanResult{
					Type:   msg.ScanType,
					Result: msg.Result,
				})
				fileExists = true
				break
			}
		}
		if !fileExists {
			currentDir.Files = append(currentDir.Files, models.File{
				Path: msg.FilePath,
				ScanResults: []models.ScanResult{
					{
						Type:   msg.ScanType,
						Result: msg.Result,
					},
				},
			})
		}
	}

	var results []models.ScanReport
	for _, report := range scanMap {
		results = append(results, *report)
	}
	return results
}

// Разбиваем путь на абсолютный, список директорий и имя файла
func splitPath(filePath string) (string, []string, string) {
	parts := strings.Split(filePath, "/")
	if len(parts) < 2 {
		return filePath, nil, ""
	}
	root := "/" + parts[1]          // Первым элементом должен быть корневой каталог
	fileName := parts[len(parts)-1] // Имя файла
	if len(parts) > 2 {
		return root, parts[1 : len(parts)-1], fileName // Полные поддиректории без имени файла
	}
	return root, nil, fileName
}
