package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	directorylisterservice "ftp-scanner_try2/internal/directory-lister-service"
	ftpclient "ftp-scanner_try2/internal/ftp"
	"ftp-scanner_try2/internal/kafka"
)

func main() {

	// Создаем Kafka consumer и producer
	consumer := kafka.NewConsumer([]string{"localhost:9092"}, "directories-to-scan", "directory-lister-group")
	defer consumer.CloseReader()

	producer, err := kafka.NewProducer("localhost:9092")
	if err != nil {
		log.Fatal("Ошибка создания Kafka producer:", err)
	}
	defer producer.CloseWriter()

	// Подключение к FTP серверу
	ftpClient, err := ftpclient.NewFTPClient("127.0.0.1:21", "user", "pass")
	if err != nil {
		log.Fatal("Ошибка подключения к FTP:", err)
	}
	defer ftpClient.Close()
	// Создаем репозиторий и сервис
	service := directorylisterservice.NewDirectoryListerService(ftpClient, producer)
	// Создаем Kafka kafkaHandler
	kafkaHandler := directorylisterservice.NewKafkaHandler(service, consumer)

	// Канал для graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск обработки сообщений
	ctx, cancel := context.WithCancel(context.Background())
	go kafkaHandler.Start(ctx)

	// Ожидание сигнала завершения
	<-stop
	cancel()
	log.Println("✅ Сервис завершил работу")
}
