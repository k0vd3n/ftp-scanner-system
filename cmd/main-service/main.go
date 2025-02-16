package main

import (
	"log"
	"net/http"
	"os"

	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/internal/kafka"
	mainservice "ftp-scanner_try2/internal/main-service"
	service "ftp-scanner_try2/internal/main-service/services"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	grpcReportServerAddress := os.Getenv("GRPC_REPORT_SERVER_ADDRESS")
	grpcCounterServerAddress := os.Getenv("GRPC_COUNTER_SERVER_ADDRESS")

	// Инициализация Kafka Producer
	kafkaProducer, err := kafka.NewProducer("kafka:9092") // Адрес Kafka брокера
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.CloseWriter()

	// Создаем gRPC соединения
	creds := insecure.NewCredentials()

	reportConn, err := grpc.NewClient(grpcReportServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect to Report Service: %v", err)
	}
	defer reportConn.Close()

	counterConn, err := grpc.NewClient(grpcCounterServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect to Counter Service: %v", err)
	}
	defer counterConn.Close()

	// Создаем сервисы
	scanService := service.NewKafkaScanService(kafkaProducer)
	reportService := service.NewGRPCReportService(proto.NewReportServiceClient(reportConn))
	counterService := service.NewGRPCCounterService(proto.NewCounterServiceClient(counterConn))

	// Создаем MainServer
	server := mainservice.NewMainServer(scanService, reportService, counterService)

	// Настройка роутера
	router := server.SetupRouter()

	// Запуск HTTP сервера
	log.Println("Main Service is running on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
