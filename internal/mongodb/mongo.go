package mongodb

import (
	"context"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/models"
	"log"

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

func NewMongoReportRepository(client *mongo.Client, database, collection string) ReportRepository {
	return &MongoReportRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}

func (r *MongoReportRepository) GetReportsByScanID(ctx context.Context, scanID string) ([]models.ScanReport, error) {
	log.Printf("MongoDB: Получение отчетов по ID скана: %s\n", scanID)
	collection := r.client.Database(r.database).Collection(r.collection)
	filter := bson.M{"scan_id": scanID}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Printf("MongoDB: Ошибка при получении отчетов: %v\n", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var reports []models.ScanReport
	for cursor.Next(ctx) {
		var report models.ScanReport
		if err := cursor.Decode(&report); err != nil {
			log.Printf("MongoDB: Ошибка при декодировании отчета: %v\n", err)
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
func NewMongoCounterRepository(client *mongo.Client, database string) GetCounterRepository {
	return &mongoCounterRepository{
		client:   client,
		database: database,
	}
}

// Метод получения счетчиков
func (r *mongoCounterRepository) GetCountersByScanID(ctx context.Context, scanID string, config config.MongoCounterSvcConfig) (*models.CounterResponseGRPC, error) {
	log.Printf("MongoDB: Получение счетчиков для ID скана: %s\n", scanID)
	err := godotenv.Load()
	if err != nil {
		log.Printf("MongoDB: Ошибка при загрузке переменных окружения: %v\n", err)
		log.Fatalf("Error loading .env file: %v", err)
	}

	counters := &models.CounterResponseGRPC{ScanID: scanID}

	type counterMapping struct {
		CollectionName string
		TargetField    *int64
	}

	mappings := []counterMapping{
		{config.ScanDirectoriesCount, &counters.DirectoriesCount},
		{config.ScanFilesCount, &counters.FilesCount},
		{config.CompletedDirectoriesCount, &counters.CompletedDirectories},
		{config.CompletedFilesCount, &counters.CompletedFiles},
	}

	for _, mapping := range mappings {
		collection := r.client.Database(r.database).Collection(mapping.CollectionName)
		filter := bson.M{"scan_id": scanID}

		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			log.Printf("MongoDB: Ошибка при получении счетчиков: %v\n", err)
			return nil, err
		}
		defer cursor.Close(ctx)

		var total int64
		for cursor.Next(ctx) {
			var data struct {
				Number int64 `bson:"number"`
			}
			if err := cursor.Decode(&data); err != nil {
				log.Printf("MongoDB: Ошибка при декодировании счетчика: %v\n", err)
				return nil, err
			}
			total += data.Number
		}

		*mapping.TargetField = total
	}

	log.Printf("MongoDB: Успешное получение счетчиков для ID скана: %s\n", scanID)
	return counters, nil
}

type MongoCounterRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

func NewCounterRepository(client *mongo.Client, database, collection string) CounterReducerRepository {
	return &MongoCounterRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}

func (r *MongoCounterRepository) InsertReducedCounters(ctx context.Context, counts []models.CountMessage) error {
	log.Printf("MongoDB: Вставка счетчиков в коллекцию %s\n", r.collection)
	collection := r.client.Database(r.database).Collection(r.collection)

	var documents []interface{}
	for _, count := range counts {
		documents = append(documents, count)
	}

	_, err := collection.InsertMany(ctx, documents)
	if err != nil {
		log.Printf("MongoDB: Ошибка вставки данных в коллекцию %s: %v\n", r.collection, err)
		return err
	}

	log.Printf("MongoDB: Успешно вставлено %d документов в коллекцию %s\n", len(documents), r.collection)
	return nil
}

type MongoSaveReportRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

func NewMongoSaveReportRepository(client *mongo.Client, database, collection string) SaveReportRepository {
	return &MongoSaveReportRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}

func (r *MongoSaveReportRepository) InsertScanReports(ctx context.Context, reports []models.ScanReport) error {
	log.Printf("MongoDB: Вставка отчетов в коллекцию %s\n", r.collection)
	collection := r.client.Database(r.database).Collection(r.collection)

	var docs []interface{}
	for _, report := range reports {
		docs = append(docs, report)
	}

	_, err := collection.InsertMany(ctx, docs)
	if err != nil {
		log.Printf("MongoDB: Ошибка при вставке отчетов: %v", err)
		return err
	}

	log.Printf("MongoDB: Успешно вставлено %d отчетов", len(docs))
	return nil
}