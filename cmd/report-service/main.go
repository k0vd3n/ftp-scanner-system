package main

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/internal/mongodb"
	reportservice "ftp-scanner_try2/internal/scan-reports-service"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	log.Printf("Report Service: main: Запуск Report Service...")
	log.Printf("Report Service: main: Загрузка переменных окружения...")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Report Service: main: Ошибка загрузки .env: %v", err)
	}
	mongoURI := os.Getenv("MONGO_URI")
	mongoCollection := os.Getenv("MONGO_COLLECTION_REPORTS")
	mongoDatabase := os.Getenv("MONGO_DATABASE_REPORTS")
	serverPort := os.Getenv("GRPC_REPORT_SERVER_PORT")

	log.Printf("Report Service: main: Инициализация соединения с MongoDB...")
	log.Printf("Report Service: main: MongoDB URI: %s", mongoURI)
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Report Service: main: Ошибка подключения к MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	log.Printf("Report Service: main: Инициализация сервисов...")
	repo := mongodb.NewMongoReportRepository(client, mongoDatabase, mongoCollection)
	storage := reportservice.NewFileReportStorage("reports")

	service := reportservice.NewReportService(repo, storage)
	server := reportservice.NewReportServer(service)

	log.Printf("Report Service: main: Инициализация gRPC соединений...")
	grpcServer := grpc.NewServer()
	proto.RegisterReportServiceServer(grpcServer, server)

	lis, err := net.Listen("tcp", serverPort)
	if err != nil {
		log.Fatalf("Report Service: main: Ошибка создания TCP-сокета: %v", err)
	}
	log.Printf("Report Service: main: report-service запущен на порту %s", serverPort)
	grpcServer.Serve(lis)
}
