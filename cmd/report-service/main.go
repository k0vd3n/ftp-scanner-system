package main

/*
import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	reportservice "ftp-scanner_try2/internal/report-service"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	mongoURI := os.Getenv("MONGODB_URI")
	reportServicePort := os.Getenv("REPORT_SERVICE_PORT")
	// Инициализация MongoDB
	server, err := reportservice.NewReportServer(mongoURI)
	if err != nil {
		log.Fatalf("Failed to initialize ReportServer: %v", err)
	}
	defer func() {
		if err := server.MongoClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("Failed to disconnect MongoDB client: %v", err)
		}
	}()

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", reportServicePort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterReportServiceServer(grpcServer, server)

	log.Printf("Report Service is running on port %s", reportServicePort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
*/

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/internal/mongodb"
	reportservice "ftp-scanner_try2/internal/scan-reports-service"
	"log"
	"net"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	mongoCollection := os.Getenv("MONGO_COLLECTION_REPORTS")
	mongoDatabase := os.Getenv("MONGO_DATABASE_REPORTS")

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	repo := mongodb.NewMongoReportRepository(client, mongoDatabase, mongoCollection)
	storage := reportservice.NewFileReportStorage("reports")

	service := reportservice.NewReportService(repo, storage)
	server := reportservice.NewReportServer(service)

	grpcServer := grpc.NewServer()
	proto.RegisterReportServiceServer(grpcServer, server)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Println("Report Service is running on port 50051")
	grpcServer.Serve(lis)
}
