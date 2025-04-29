package mongodb

import (
	"context"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Реализация интерфейса для
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
func (r *mongoCounterRepository) GetCountersByScanID(ctx context.Context, scanID string, config config.StatusServiceMongo) (*models.StatusResponseGRPC, error) {
	log.Printf("MongoDB: Получение счетчиков для ID скана: %s\n", scanID)

	counters := &models.StatusResponseGRPC{ScanID: scanID}

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

	log.Printf("MongoDB: Получение счетчиков из коллекций %s, %s, %s, %s для scan_id=%s\n",
		config.ScanDirectoriesCount, config.ScanFilesCount, config.CompletedDirectoriesCount, config.CompletedFilesCount, scanID)

	for _, mapping := range mappings {
		collection := r.client.Database(r.database).Collection(mapping.CollectionName)
		log.Printf("MongoDB: Получение счетчиков из коллекции %s для scan_id: %s\n", mapping.CollectionName, scanID)
		filter := bson.M{"scanid": scanID}

		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			log.Printf("MongoDB: Ошибка при получении счетчиков: %v\n", err)
			return nil, err
		}
		log.Printf("MongoDB: Получены счетчики из коллекции %s для scan_id: %s\n", mapping.CollectionName, scanID)
		defer cursor.Close(ctx)

		var total int64
		counter := 0
		for cursor.Next(ctx) {
			var data struct {
				Number int64 `bson:"number"`
			}
			if err := cursor.Decode(&data); err != nil {
				log.Printf("MongoDB: Ошибка при декодировании счетчика: %v\n", err)
				return nil, err
			}
			if counter != 10 {
				log.Printf("MongoDB: Декодировано число %d счетчика из коллекции %s для scan_id: %s\n", data.Number, mapping.CollectionName, scanID)
			}
			total += data.Number
		}
		log.Printf("MongoDB: Итоговое число счетчика = %d из коллекции %s для scan_id: %s\n", total, mapping.CollectionName, scanID)

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
	log.Printf("MongoDB: Вставляем %d документов в коллекцию %s\n", len(documents), r.collection)

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

// MongoMetricRepository — MongoDB-реализация
type MongoMetricRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

func NewMongoMetricRepository(client *mongo.Client, db, coll string) MetricRepository {
	return &MongoMetricRepository{client, db, coll}
}

func (r *MongoMetricRepository) Save(ctx context.Context, instance string, payload []byte) error {
	col := r.client.Database(r.database).Collection(r.collection)
	doc := bson.M{
		"instance": instance,
		"payload":  payload,
		"saved_at": time.Now(),
	}
	_, err := col.InsertOne(ctx, doc)
	return err
}

func (r *MongoMetricRepository) Load(ctx context.Context, instance string) ([]byte, error) {
	col := r.client.Database(r.database).Collection(r.collection)
	filter := bson.M{"instance": instance}
	opts := options.FindOne().SetSort(bson.M{"saved_at": -1})
	var result struct {
		Payload []byte `bson:"payload"`
	}
	err := col.FindOne(ctx, filter, opts).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return result.Payload, err
}
