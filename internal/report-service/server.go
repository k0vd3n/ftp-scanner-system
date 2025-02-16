/*
package reportservice

import (
	"context"
	"encoding/json"
	"fmt"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/internal/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReportServer struct {
	proto.UnimplementedReportServiceServer
	MongoClient *mongo.Client
}

func NewReportServer(mongoURI string) (*ReportServer, error) {
	// Инициализация MongoDB клиента
	ctx := context.TODO()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Проверка подключения
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB")

	return &ReportServer{
		MongoClient: client,
	}, nil
}

func (s *ReportServer) GenerateReport(ctx context.Context, req *proto.ReportRequest) (*proto.ReportResponse, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	database := os.Getenv("MONGODB_DATABASE")
	collection := os.Getenv("MONGODB_COLLECTION")

	scanID := req.GetScanId()
	log.Printf("Generating report for scan_id: %s", scanID)

	// Получение данных из MongoDB
	reportCollection := s.MongoClient.Database(database).Collection(collection)
	filter := bson.M{"scan_id": scanID} // Используем scan_id из запроса

	cursor, err := reportCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query MongoDB: %v", err)
	}
	defer cursor.Close(ctx)

	// Чтение данных
	var reports []models.ScanReport
	for cursor.Next(ctx) {
		var report models.ScanReport
		if err := cursor.Decode(&report); err != nil {
			return nil, fmt.Errorf("failed to decode MongoDB document: %v", err)
		}
		reports = append(reports, report)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	// Группировка данных
	groupedData := groupScanResults(reports)

	// Сохранение отчета в файл
	reportFileName := fmt.Sprintf("reports/%s.json", scanID)
	if err := saveReportToFile(groupedData, reportFileName); err != nil {
		return nil, fmt.Errorf("failed to save report to file: %v", err)
	}

	// Придумать как получить ссылку на отчет
	// Возврат ссылки на отчет
	reportURL := fmt.Sprintf("http://example.com/reports/%s.json", scanID)
	return &proto.ReportResponse{
		ScanId:    scanID,
		ReportUrl: reportURL,
	}, nil
}

// Группировка результатов сканирования
func groupScanResults(reports []models.ScanReport) []models.ScanReport {
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

	// Преобразуем map в слайс
	var groupedResults []models.ScanReport
	for _, report := range scanMap {
		groupedResults = append(groupedResults, *report)
	}
	return groupedResults
}

// Сохранение отчета в файл
func saveReportToFile(reports []models.ScanReport, fileName string) error {
	// Создаем директорию, если она не существует
	if err := os.MkdirAll("reports", 0755); err != nil {
		return err
	}

	// Сериализация данных в JSON
	jsonData, err := json.MarshalIndent(reports, "", "  ")
	if err != nil {
		return err
	}

	// Запись в файл
	if err := os.WriteFile(fileName, jsonData, 0644); err != nil {
		return err
	}

	log.Printf("Report saved to file: %s", fileName)
	return nil
}
*/

package reportservice

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
)

type ReportServer struct {
	proto.UnimplementedReportServiceServer
	service ReportServiceInterface
}

func NewReportServer(service *ReportService) *ReportServer {
	return &ReportServer{service: service}
}

func (s *ReportServer) GenerateReport(ctx context.Context, req *proto.ReportRequest) (*proto.ReportResponse, error) {
	reportURL, err := s.service.GenerateReport(ctx, req.GetScanId())
	if err != nil {
		return nil, err
	}

	return &proto.ReportResponse{
		ScanId:    req.GetScanId(),
		ReportUrl: reportURL,
	}, nil
}
