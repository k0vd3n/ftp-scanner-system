package mongodb

import (
	"context"
	"ftp-scanner_try2/internal/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Реализация интерфейса для MongoDB
type MongoReportRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

func NewMongoReportRepository(client *mongo.Client, database, collection string) *MongoReportRepository {
	return &MongoReportRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}

func (r *MongoReportRepository) GetReportsByScanID(ctx context.Context, scanID string) ([]models.ScanReport, error) {
	collection := r.client.Database(r.database).Collection(r.collection)
	filter := bson.M{"scan_id": scanID}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reports []models.ScanReport
	for cursor.Next(ctx) {
		var report models.ScanReport
		if err := cursor.Decode(&report); err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	return reports, cursor.Err()
}

type mongoCounterRepository struct {
	client   *mongo.Client
	database string
}

// Конструктор репозитория
func NewMongoCounterRepository(client *mongo.Client, database string) CounterRepository {
	return &mongoCounterRepository{
		client:   client,
		database: database,
	}
}

// Метод получения счетчиков
func (r *mongoCounterRepository) GetCountersByScanID(ctx context.Context, scanID string) (*models.CounterResponseGRPC, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	scanDirectoriesCount := os.Getenv("SCAN_DIRECTORIES_COUNT")
	scanFilesCount := os.Getenv("SCAN_FILES_COUNT")
	completedDirectoriesCount := os.Getenv("COMPLETED_DIRECTORIES_COUNT")
	completedFilesCount := os.Getenv("COMPLETED_FILES_COUNT")

	counters := &models.CounterResponseGRPC{ScanID: scanID}

	type counterMapping struct {
		CollectionName string
		TargetField    *int64
	}

	mappings := []counterMapping{
		{scanDirectoriesCount, &counters.DirectoriesCount},
		{scanFilesCount, &counters.FilesCount},
		{completedDirectoriesCount, &counters.CompletedDirectories},
		{completedFilesCount, &counters.CompletedFiles},
	}

	for _, mapping := range mappings {
		collection := r.client.Database(r.database).Collection(mapping.CollectionName)
		filter := bson.M{"scan_id": scanID}

		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		var total int64
		for cursor.Next(ctx) {
			var data struct {
				Number int64 `bson:"number"`
			}
			if err := cursor.Decode(&data); err != nil {
				return nil, err
			}
			total += data.Number
		}

		*mapping.TargetField = total
	}

	return counters, nil
}
