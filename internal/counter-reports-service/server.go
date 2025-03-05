package counterservice

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/config"
	"log"
)

/*
import (
	"context"
	"fmt"
	"ftp-scanner_try2/api/grpc/proto"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CounterServer struct {
	proto.UnimplementedCounterServiceServer
	MongoClient *mongo.Client
}

func NewCounterServer(mongoURI string) (*CounterServer, error) {
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

	return &CounterServer{
		MongoClient: client,
	}, nil
}

func (s *CounterServer) GetCounters(ctx context.Context, req *proto.CounterRequest) (*proto.CounterResponse, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	database := os.Getenv("MONGODB_DATABASE")
	scanID := req.GetScanId()
	log.Printf("Fetching counters for scan_id: %s", scanID)

	// Коллекции для счетчиков
	collections := map[string]string{
		"directories_count":     "scan_directories_count",
		"files_count":           "scan_files_count",
		"completed_directories": "completed_directories_count",
		"completed_files":       "completed_files_count",
	}

	// Результаты счетчиков (используем int64)
	counters := make(map[string]int64)

	// Запрос данных из MongoDB
	for key, collectionName := range collections {
		collection := s.MongoClient.Database(database).Collection(collectionName)
		filter := bson.M{"scan_id": scanID}

		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to query MongoDB collection %s: %v", collectionName, err)
		}
		defer cursor.Close(ctx)

		// Суммирование значений
		var total int64 = 0
		for cursor.Next(ctx) {
			var data struct {
				Number int64 `bson:"number"` // Изменили тип на int64
			}
			if err := cursor.Decode(&data); err != nil {
				return nil, fmt.Errorf("failed to decode MongoDB document: %v", err)
			}
			total += data.Number
		}

		if err := cursor.Err(); err != nil {
			return nil, fmt.Errorf("cursor error: %v", err)
		}

		counters[key] = total
	}

	// Возврат результата
	return &proto.CounterResponse{
		ScanId:               scanID,
		DirectoriesCount:     counters["directories_count"],
		FilesCount:           counters["files_count"],
		CompletedDirectories: counters["completed_directories"],
		CompletedFiles:       counters["completed_files"],
	}, nil
}
*/

type CounterServerInterface interface {
	proto.CounterServiceServer
}

// структура для gRPC сервера
type CounterServer struct {
	proto.UnimplementedCounterServiceServer
	service CounterService
	config  config.MongoCounterSvcConfig
}

// Конструктор сервера
func NewCounterServer(service CounterService, config config.MongoCounterSvcConfig) CounterServerInterface {
	return &CounterServer{
		service: service,
		config:  config,
	}
}

// gRPC метод получения счетчиков
func (s *CounterServer) GetCounters(ctx context.Context, req *proto.CounterRequest) (*proto.CounterResponse, error) {
	log.Printf("Получаем счетчики для scan_id: %s", req.ScanId)

	counters, err := s.service.GetCounters(ctx, req.ScanId, s.config)
	if err != nil {
		log.Printf("Ошибка получения счетчиков для scan_id %s: %v", req.ScanId, err)
		return nil, err
	}

	return &proto.CounterResponse{
		ScanId:               counters.ScanID,
		DirectoriesCount:     counters.DirectoriesCount,
		FilesCount:           counters.FilesCount,
		CompletedDirectories: counters.CompletedDirectories,
		CompletedFiles:       counters.CompletedFiles,
	}, nil
}
