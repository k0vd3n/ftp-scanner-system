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

		// Если директории нет, создаем ее с абсолютным путем
		if currentDir == nil {
			newDir := models.Directory{
				Directory:    fullPath,
				Subdirectory: []models.Directory{},
				Files:        []models.File{},
			}
			currentReport.Directories = append(currentReport.Directories, newDir)
			currentDir = &currentReport.Directories[len(currentReport.Directories)-1]
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
		}
	}

	// Преобразуем map в слайс для JSON-ответа
	var groupedResults []models.ScanReport
	for _, report := range scanMap {
		groupedResults = append(groupedResults, *report)
	}
	return groupedResults
}

// Разбиваем путь на абсолютный путь, список директорий и имя файла
func splitPath(filePath string) (string, []string, string) {
	parts := strings.Split(filePath, "/")
	if len(parts) < 2 {
		return filePath, nil, ""
	}
	root := "/" + parts[1]           // Корневой каталог
	fileName := parts[len(parts)-1]   // Имя файла

	// Если есть поддиректории, исключаем дублирование корневой директории
	if len(parts) > 2 {
		// parts[1] соответствует корневой директории, поэтому начинаем с parts[2]
		return root, parts[2 : len(parts)-1], fileName
	}
	return root, nil, fileName
}
