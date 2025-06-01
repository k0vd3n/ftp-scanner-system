package mongodb

import (
	"context"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Реализация интерфейса для
type mongoGenerateReportRepository struct {
	client      *mongo.Client
	database    string
	collection  string
	collection2 string
	logger      *zap.Logger
}

func NewMongoReportRepository(client *mongo.Client, database, collection, collection2 string, logger *zap.Logger) GenerateReportRepository {
	return &mongoGenerateReportRepository{
		client:      client,
		database:    database,
		collection:  collection,
		collection2: collection2,
		logger:      logger,
	}
}

func (r *mongoGenerateReportRepository) GetReportsByScanID(ctx context.Context, scanID string) ([]models.ScanReport, error) {
	r.logger.Info("MongoDB: Получение отчетов по ID скана", zap.String("scanID", scanID))
	collection := r.client.Database(r.database).Collection(r.collection)
	filter := bson.M{"scan_id": scanID}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		r.logger.Error("MongoDB: Ошибка при получении отчетов", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var reports []models.ScanReport
	for cursor.Next(ctx) {
		var report models.ScanReport
		if err := cursor.Decode(&report); err != nil {
			r.logger.Error("MongoDB: Ошибка при декодировании отчета", zap.Error(err))
			return nil, err
		}
		reports = append(reports, report)
	}

	return reports, cursor.Err()
}

func (r *mongoGenerateReportRepository) SaveResult(ctx context.Context, report models.ScanReport) error {
	r.logger.Info("MongoDB: Сохранение итогового отчета в базу данных")
	collection := r.client.Database(r.database).Collection(r.collection2)
	r.logger.Info("MongoDB: Сохранение в коллекцию", zap.String("collection", r.collection2))

	_, err := collection.InsertOne(ctx, report)
	if err != nil {
		r.logger.Error("MongoDB: Ошибка при сохранении отчета", zap.Error(err))
		return err
	}
	r.logger.Info("MongoDB: Отчет успешно сохранен")
	return nil
}

type mongoGetReportRepository struct {
	client     *mongo.Client
	database   string
	collection string
	logger     *zap.Logger
}

func NewMongoGetReportRepository(client *mongo.Client, database, collection string, logger *zap.Logger) GetReportRepository {
	return &mongoGetReportRepository{
		client:     client,
		database:   database,
		collection: collection,
		logger:     logger,
	}
}

func (r *mongoGetReportRepository) GetReport(ctx context.Context, scanID string) (*models.ScanReport, error) {
	r.logger.Info("MongoDB: Получение отчета по ID скана", zap.String("scanID", scanID))
	collection := r.client.Database(r.database).Collection(r.collection)
	filter := bson.M{"scan_id": scanID}
	findOpts := options.FindOne().
		SetSort(bson.D{bson.E{Key: "created_at", Value: -1}})

	var result models.ScanReport
	err := collection.FindOne(ctx, filter, findOpts).Decode(&result)
	if err == mongo.ErrNoDocuments {
		r.logger.Info("MongoDB: Отчет не найден", zap.String("scanID", scanID))
		return nil, err
	} else if err != nil {
		r.logger.Error("MongoDB: Ошибка при получении отчета", zap.Error(err))
		return nil, err
	}
	r.logger.Info("MongoDB: Отчет успешно получен", zap.String("scanID", scanID))
	return &result, nil

}

type mongoCounterRepository struct {
	client   *mongo.Client
	database string
	logger   *zap.Logger
}

// Конструктор репозитория
func NewMongoCounterRepository(client *mongo.Client, database string, logger *zap.Logger) GetCounterRepository {
	return &mongoCounterRepository{
		client:   client,
		database: database,
		logger:   logger,
	}
}

// Метод получения счетчиков
func (r *mongoCounterRepository) GetCountersByScanID(ctx context.Context, scanID string, config config.StatusServiceMongo) (*models.StatusResponseGRPC, error) {
	r.logger.Info("MongoDB: Получение счетчиков по ID скана", zap.String("scanID", scanID))

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

	r.logger.Info("MongoDB: Получение счетчиков по ID скана из коллекций",
		zap.String("scanID", scanID),
		zap.String("collections", config.ScanDirectoriesCount),
		zap.String("collections", config.ScanFilesCount),
		zap.String("collections", config.CompletedDirectoriesCount),
		zap.String("collections", config.CompletedFilesCount))

	for _, mapping := range mappings {
		collection := r.client.Database(r.database).Collection(mapping.CollectionName)
		r.logger.Info("MongoDB: Получение счетчиков из коллекции для ID скана",
			zap.String("collection", mapping.CollectionName),
			zap.String("scanID", scanID))
		filter := bson.M{"scanid": scanID}

		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			r.logger.Error("MongoDB: Ошибка при получении счетчиков", zap.Error(err))
			return nil, err
		}
		r.logger.Info("MongoDB: Получение счетчиков из коллекции", zap.String("collection", mapping.CollectionName), zap.String("scanID", scanID))
		defer cursor.Close(ctx)

		var total int64
		counter := 0
		for cursor.Next(ctx) {
			var data struct {
				Number int64 `bson:"number"`
			}
			if err := cursor.Decode(&data); err != nil {
				r.logger.Error("MongoDB: Ошибка при декодировании счетчика", zap.Error(err))
				return nil, err
			}
			if counter != 10 {
				r.logger.Info("MongoDB: Декодировано число счетчика из коллекции",
					zap.String("collection", mapping.CollectionName),
					zap.String("scanID", scanID),
					zap.Int64("number", data.Number))
			}
			total += data.Number
		}
		r.logger.Info("MongoDB: Итоговое число счетчика из коллекции",
			zap.String("collection", mapping.CollectionName),
			zap.String("scanID", scanID),
			zap.Int64("total", total))

		*mapping.TargetField = total
	}

	r.logger.Info("MongoDB: Счетчики успешно получены", zap.String("scanID", scanID))
	return counters, nil
}

type MongoCounterRepository struct {
	client     *mongo.Client
	database   string
	collection string
	logger     *zap.Logger
}

func NewCounterRepository(client *mongo.Client, database, collection string, logger *zap.Logger) CounterReducerRepository {
	return &MongoCounterRepository{
		client:     client,
		database:   database,
		collection: collection,
		logger:     logger,
	}
}

func (r *MongoCounterRepository) InsertReducedCounters(ctx context.Context, counts []models.CountMessage) error {
	r.logger.Info("MongoDB: Вставка счетчиков в коллекцию", zap.String("collection", r.collection))
	collection := r.client.Database(r.database).Collection(r.collection)

	var documents []interface{}
	for _, count := range counts {
		documents = append(documents, count)
	}
	r.logger.Info("MongoDB: Вставка счетчиков в коллекцию",
		zap.String("collection", r.collection),
		zap.Int("count", len(documents)))

	_, err := collection.InsertMany(ctx, documents)
	if err != nil {
		r.logger.Error("MongoDB: Ошибка вставки данных в коллекцию",
			zap.String("collection", r.collection),
			zap.Error(err))
		return err
	}

	r.logger.Info("MongoDB: Успешно вставлено документов в коллекцию",
		zap.String("collection", r.collection),
		zap.Int("count", len(documents)))
	return nil
}

type MongoSaveReportRepository struct {
	client     *mongo.Client
	database   string
	collection string
	logger     *zap.Logger
}

func NewMongoSaveReportRepository(client *mongo.Client, database, collection string, logger *zap.Logger) SaveReportRepository {
	return &MongoSaveReportRepository{
		client:     client,
		database:   database,
		collection: collection,
		logger:     logger,
	}
}

func (r *MongoSaveReportRepository) InsertScanReports(ctx context.Context, reports []models.ScanReport) error {
	r.logger.Info("MongoDB: Вставка отчетов в коллекцию", zap.String("collection", r.collection))
	collection := r.client.Database(r.database).Collection(r.collection)

	var docs []interface{}
	for _, report := range reports {
		docs = append(docs, report)
	}

	_, err := collection.InsertMany(ctx, docs)
	if err != nil {
		r.logger.Error("MongoDB: Ошибка при вставке отчетов", zap.Error(err))
		return err
	}

	r.logger.Info("MongoDB: Успешно вставлено отчетов", zap.Int("count", len(docs)))
	return nil
}

// MongoMetricRepository — MongoDB-реализация
type MongoMetricRepository struct {
	client     *mongo.Client
	database   string
	collection string
	logger     *zap.Logger
}

func NewMongoMetricRepository(client *mongo.Client, db, coll string, logger *zap.Logger) MetricRepository {
	return &MongoMetricRepository{client, db, coll, logger}
}

func (r *MongoMetricRepository) Save(ctx context.Context, instance string, payload []byte) error {
	r.logger.Info("MongoDB: Сохранение метрики в коллекцию", zap.String("collection", r.collection))
	col := r.client.Database(r.database).Collection(r.collection)
	doc := bson.M{
		"instance": instance,
		"payload":  payload,
		"saved_at": time.Now(),
	}
	_, err := col.InsertOne(ctx, doc)
	if err != nil {
		r.logger.Error("MongoDB: Ошибка при сохранении метрики", zap.Error(err))
	}
	r.logger.Info("MongoDB: Метрика успешно сохранена в коллекцию", zap.String("collection", r.collection))
	return err
}

func (r *MongoMetricRepository) Load(ctx context.Context, instance string) ([]byte, error) {
	r.logger.Info("MongoDB: Загрузка метрики из коллекции", zap.String("collection", r.collection))
	col := r.client.Database(r.database).Collection(r.collection)
	filter := bson.M{"instance": instance}
	opts := options.FindOne().SetSort(bson.M{"saved_at": -1})
	var result struct {
		Payload []byte `bson:"payload"`
	}
	r.logger.Info("MongoDB: Поиск метрики в коллекции", zap.String("collection", r.collection))
	err := col.FindOne(ctx, filter, opts).Decode(&result)
	if err == mongo.ErrNoDocuments {
		r.logger.Info("MongoDB: Метрика не найдена в коллекции", zap.String("collection", r.collection))
		return nil, nil
	}
	r.logger.Info("MongoDB: Метрика успешно загружена из коллекции", zap.String("collection", r.collection))
	return result.Payload, err
}
