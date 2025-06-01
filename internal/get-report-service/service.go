package getreportservice

import (
	"context"
	"fmt"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/internal/models"
	"ftp-scanner_try2/internal/mongodb"
	"log"
	"strings"

	"go.uber.org/zap"
)

type GetReportServiceInterface interface {
	GetDirectory(ctx context.Context, request *proto.DirectoryRequest) (*proto.DirectoryResponse, error)
}

type GetReportService struct {
	repo   mongodb.GetReportRepository
	logger *zap.Logger
}

func NewGetReportService(repo mongodb.GetReportRepository, logger *zap.Logger) *GetReportService {
	return &GetReportService{repo: repo, logger: logger}
}

func (s *GetReportService) GetDirectory(ctx context.Context, req *proto.DirectoryRequest) (*proto.DirectoryResponse, error) {
	scanID := req.GetScanId()
	s.logger.Info("GetReportService: GetDirectory: Получение директории", zap.String("scan_id", scanID))
	report, err := s.repo.GetReport(ctx, scanID)
	if err != nil {
		s.logger.Error("GetReportService: GetDirectory: Ошибка получения отчета для scan_id="+scanID, zap.Error(err))
		return nil, err
	}

	log.Printf("GetReportService: GetDirectory: Отчет получен для scan_id=%s", scanID)
	var targetDir *models.Directory
	if req.GetPath() == "" {
		if len(report.Directories) == 0 {
			s.logger.Info("GetReportService: GetDirectory: Нет директорий в отчете для scan_id="+scanID)
			return nil, fmt.Errorf("нет директорий в отчете для scan_id=%s", scanID)
		}
		targetDir = &report.Directories[0]
	} else {
		s.logger.Info("GetReportService: GetDirectory: Поиск директории по пути", zap.String("path", req.GetPath()))
		targetDir, err = findDirectoryByPath(report.Directories, req.GetPath())
		if err != nil {
			s.logger.Error("GetReportService: GetDirectory: Ошибка поиска директории по пути", zap.String("path", req.GetPath()), zap.String("scan_id", scanID), zap.Error(err))
			return nil, err
		}
	}

	s.logger.Info("GetReportService: GetDirectory: Директория получена для scan_id="+scanID)
	resp := &proto.DirectoryResponse{
		ScanId: scanID,
		Directory: &proto.Directory{
			Directory:    targetDir.Directory,
			Subdirectory: convertDirs(targetDir.Subdirectory),
			Files:        convertFiles(targetDir.Files),
		},
	}
	return resp, nil
}

// findDirectoryByPath — точно такая же реализация, как у вас была, только без loadJSON
func findDirectoryByPath(dirs []models.Directory, targetPath string) (*models.Directory, error) {
	targetPath = normalizePath(targetPath)
	if targetPath == "" {
		if len(dirs) == 0 {
			return nil, fmt.Errorf("GetReportService findDirectoryByPath: no directories available")
		}
		return &dirs[0], nil
	}
	parts := splitPathParts(targetPath)
	current := dirs

	for depth := 0; depth < len(parts); depth++ {
		expected := "/" + joinPathParts(parts[:depth+1])
		var found *models.Directory
		for i := range current {
			if current[i].Directory == expected {
				found = &current[i]
				break
			}
		}
		if found == nil {
			return nil, fmt.Errorf("GetReportService findDirectoryByPath: directory not found: %s", expected)
		}
		if depth < len(parts)-1 {
			current = found.Subdirectory
		} else {
			return found, nil
		}
	}
	return nil, fmt.Errorf("GetReportService findDirectoryByPath: directory not found")
}

func normalizePath(p string) string {
	// строчка без концевого слеша и без пробелов
	return strings.TrimRight(strings.TrimSpace(p), "/")
}

func splitPathParts(p string) []string {
	return strings.Split(strings.Trim(p, "/"), "/")
}

func joinPathParts(parts []string) string {
	return strings.Join(parts, "/")
}

// convertDirs и convertFiles — точно такие же функции, как у вас, но принимают models.* и возвращают protobuf.*
func convertDirs(in []models.Directory) []*proto.Directory {
	out := make([]*proto.Directory, len(in))
	for i, d := range in {
		out[i] = &proto.Directory{
			Directory:    d.Directory,
			Subdirectory: convertDirs(d.Subdirectory),
			Files:        convertFiles(d.Files),
		}
	}
	return out
}

func convertFiles(in []models.File) []*proto.File {
	out := make([]*proto.File, len(in))
	for i, f := range in {
		results := make([]*proto.ScanResult, len(f.ScanResults))
		for j, r := range f.ScanResults {
			results[j] = &proto.ScanResult{Type: r.Type, Result: r.Result}
		}
		out[i] = &proto.File{
			Path:        f.Path,
			ScanResults: results,
		}
	}
	return out
}
