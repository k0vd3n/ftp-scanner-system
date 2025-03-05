package config

import (
	"ftp-scanner_try2/internal/file-scanner-service/scanner"
	"io/fs"
)

// type DirectoryListerConfig struct {
// 	ConsumerConfig KafkaConsumerConfig
// 	ProducerConfig KafkaDirectoryListerProducerConfig
// 	FTPConfig      FTPConfig
// }

type KafkaConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupId string
}

type DirectoryListerConfig struct {
	Broker                         string
	DirectoriesToScanTopic         string
	ScanDirectoriesCountTopic      string
	FilesToScanTopic               string
	ScanFilesCountTopic            string
	CompletedDirectoriesCountTopic string
}

type FilesScannerConfig struct {
	Broker                   string
	Permision                fs.FileMode
	ScanResultsTopic         string
	CompletedFilesCountTopic string
	PathForDownloadedFiles   string
	ScannerTypesMap          map[string]scanner.FileScanner
}

type FTPConfig struct {
	Host     string
	User     string
	Password string
}

type MongoCounterSvcConfig struct {
	ScanDirectoriesCount      string
	ScanFilesCount            string
	CompletedDirectoriesCount string
	CompletedFilesCount       string
}
