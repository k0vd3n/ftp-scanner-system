package main

import (
	"log"
	"net/http"
	"os"

	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/kafka"
	mainservice "ftp-scanner_try2/internal/main-service"
	service "ftp-scanner_try2/internal/main-service/services"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log.Printf("Main-service: main: Запуск Main-service...")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	log.Printf("Main-service: main: Загрузка переменных окружения...")

	grpcReportServerAddress := os.Getenv("GRPC_REPORT_SERVER_ADDRESS") + os.Getenv("GRPC_REPORT_SERVER_PORT")
	grpcCounterServerAddress := os.Getenv("GRPC_COUNTER_SERVER_ADDRESS") + os.Getenv("GRPC_COUNTER_SERVER_PORT")
	kafkaBroker := os.Getenv("KAFKA_DIR_LIST_SVC_BROKER")
	scanDirectoriesTopic := os.Getenv("KAFKA_DIR_LIST_SVC_TOPIC")
	mainServiceHTTTPPort := os.Getenv("MAIN_SERVICE_HTTP_PORT")



	log.Printf("Main-service: main: Инициализация Kafka Producer...")
	// Инициализация Kafka Producer
	kafkaProducer, err := kafka.NewProducer(kafkaBroker) // Адрес Kafka брокера
	if err != nil {
		log.Fatalf("Main-service: main: Ошибка инициализации Kafka Producer: %v", err)
	}
	defer kafkaProducer.CloseWriter()



	log.Printf("Main-service: main: Инициализация gRPC соединений...")
	// Создаем gRPC соединения
	creds := insecure.NewCredentials()

	reportConn, err := grpc.NewClient(grpcReportServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Main-service: main: Ошибка соединения с Report Service: %v", err)
	}
	defer reportConn.Close()

	counterConn, err := grpc.NewClient(grpcCounterServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Main-service: main: Ошибка соединения с Counter Service: %v", err)
	}
	defer counterConn.Close()



	log.Printf("Main-service: main: Инициализация сервисов...")
	// Создаем сервисы
	configScanDirTopic := config.DirectoryListerConfig{
		DirectoriesToScanTopic: scanDirectoriesTopic,
	}
	scanService := service.NewKafkaScanService(kafkaProducer, configScanDirTopic)
	reportService := service.NewGRPCReportService(proto.NewReportServiceClient(reportConn))
	counterService := service.NewGRPCCounterService(proto.NewCounterServiceClient(counterConn))



	log.Printf("Main-service: main: Инициализация MainServer...")
	// Создаем MainServer
	server := mainservice.NewMainServer(scanService, reportService, counterService)



	log.Printf("Main-service: main: Настройка роутера...")
	// Настройка роутера
	router := server.SetupRouter()

	log.Printf("Main-service: main: Запуск HTTP сервера...")
	// Запуск HTTP сервера
	log.Printf("Main-service: main: Сервер запущен на порту %s", mainServiceHTTTPPort)
	if err := http.ListenAndServe(mainServiceHTTTPPort, router); err != nil {
		log.Fatalf("Main-service: main: Ошибка запуска HTTP сервера: %v", err)
	}
}
