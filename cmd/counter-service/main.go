package main

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/config"
	counterservice "ftp-scanner_try2/internal/counter-reports-service"
	"ftp-scanner_try2/internal/mongodb"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	mongoURI := os.Getenv("MONGO_URI")
	mongoDB := os.Getenv("MONGO_DATABASE_COUNTER")
	counterServicePort := os.Getenv("COUNTER_SERVICE_PORT")
	/*
	   // Инициализация MongoDB
	   server, err := counterservice.NewCounterServer(mongoURI)

	   	if err != nil {
	   		log.Fatalf("Failed to initialize CounterServer: %v", err)
	   	}

	   	defer func() {
	   		if err := server.MongoClient.Disconnect(context.Background()); err != nil {
	   			log.Fatalf("Failed to disconnect MongoDB client: %v", err)
	   		}
	   	}()

	   // Запуск gRPC сервера
	   lis, err := net.Listen("tcp", counterServicePort)

	   	if err != nil {
	   		log.Fatalf("Failed to listen: %v", err)
	   	}

	   grpcServer := grpc.NewServer()
	   proto.RegisterCounterServiceServer(grpcServer, server)

	   log.Printf("Counter Service is running on port %s", counterServicePort)

	   	if err := grpcServer.Serve(lis); err != nil {
	   		log.Fatalf("Failed to serve: %v", err)
	   	}
	*/
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		// logger.Fatalf("Failed to connect to MongoDB: %v", err)
		log.Fatalf("Ошибка подключения к MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	repo := mongodb.NewMongoCounterRepository(client, mongoDB)
	service := counterservice.NewCounterService(repo)
	config := config.MongoCounterSvcConfig{
		ScanDirectoriesCount:      os.Getenv("SCAN_DIRECTORIES_COUNT"),
		ScanFilesCount:            os.Getenv("SCAN_FILES_COUNT"),
		CompletedDirectoriesCount: os.Getenv("COMPLETED_DIRECTORIES_COUNT"),
		CompletedFilesCount:       os.Getenv("KAFKA_COMPLETED_FILES_COUNT_TOPIC"),
	}
	server := counterservice.NewCounterServer(service, config)

	lis, err := net.Listen("tcp", counterServicePort)
	if err != nil {
		// logger.Fatalf("Failed to listen: %v", err)
		log.Fatalf("Ошибка прослушивания порта: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterCounterServiceServer(grpcServer, server)

	// logger.Infof("Counter Service is running on port %s", counterServicePort)
	log.Printf("Сервис счетчиков запущен на порте %s", counterServicePort)
	if err := grpcServer.Serve(lis); err != nil {
		// logger.Fatalf("Failed to serve: %v", err)
		log.Fatalf("Ошибка запуска сервиса: %v", err)
	}
}
