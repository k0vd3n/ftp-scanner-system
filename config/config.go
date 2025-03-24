package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

func LoadUnifiedConfig(path string) (*UnifiedConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg UnifiedConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

type UnifiedConfig struct {
	FileScanner       FileScannerConfig       `yaml:"file_scanner_service"`
	DirectoryLister   DirectoryListerConfig   `yaml:"directory_lister_service"`
	MainService       MainServiceConfig       `yaml:"main_service"`
	CounterReducer    CounterReducerConfig    `yaml:"counter_reducer_service"`
	ScanResultReducer ScanResultReducerConfig `yaml:"scan_result_reducer_service"`
	ReportService     ReportServiceConfig     `yaml:"report_service"`
}

type FileScannerConfig struct {
	KafkaConsumer                    FileScannerConsumer                    `yaml:"kafka_consumer"`
	KafkaScanResultProducer          FileScannerProducerScanResults         `yaml:"kafka_scan_result_producer"`
	KafkaCompletedFilesCountProducer FileScannerProducerCompletedFilesCount `yaml:"kafka_completed_files_count_producer"`
}

type FileScannerConsumer struct {
	Brokers       []string `yaml:"brokers"`
	ConsumerTopic string   `yaml:"consumer_topic"`
	ConsumerGroup string   `yaml:"consumer_group"`
}

type FileScannerProducerScanResults struct {
	Broker               string        `yaml:"broker"`
	FileScanDownloadPath string        `yaml:"file_scan_download_path"`
	Permision            string        `yaml:"permission"`
	ScannerTypes         []string      `yaml:"scanner_types"`
	Routing              RoutingConfig `yaml:"routing"`
}

type FileScannerProducerCompletedFilesCount struct {
	Broker                   string `yaml:"broker"`
	CompletedFilesCountTopic string `yaml:"completed_files_count_topic"`
}

type RoutingRule struct {
	ScanType     string   `yaml:"scan_type"`     // Тип сканирования (например "zero_bytes")
	TriggerValue string   `yaml:"trigger_value"` // Значение для активации правила (например "0" или "0-1024")
	OutputTopics []string `yaml:"output_topics"` // Топики для отправки
}

type RoutingConfig struct {
	Rules        []RoutingRule `yaml:"rules"`
	DefaultTopic string        `yaml:"default_topic"` // Топик по умолчанию (scan-results)
}

type DirectoryListerConfig struct {
	Kafka DirectoryListerKafkaConfig `yaml:"kafka"`
}

type DirectoryListerKafkaConfig struct {
	Broker                         string `yaml:"broker"`
	ConsumerTopic                  string `yaml:"consumer_topic"`
	ConsumerGroup                  string `yaml:"consumer_group"`
	DirectoriesToScanTopic         string `yaml:"directories_to_scan_topic"`
	ScanDirectoriesCountTopic      string `yaml:"scan_directories_count_topic"`
	FilesToScanTopic               string `yaml:"files_to_scan_topic"`
	ScanFilesCountTopic            string `yaml:"scan_files_count_topic"`
	CompletedDirectoriesCountTopic string `yaml:"completed_directories_count_topic"`
}

type MainServiceConfig struct {
	Kafka MainServiceKafka `yaml:"kafka"`
	GRPC  MainServiceGrpc  `yaml:"grpc"`
	HTTP  MainServiceHttp  `yaml:"http"`
}

type MainServiceKafka struct {
	Broker          string `yaml:"broker"`
	DirectoryTorpic string `yaml:"directory_topic"`
}

type MainServiceGrpc struct {
	ReportServerAddress  string `yaml:"report_server_address"`
	ReportServerPort     string `yaml:"report_server_port"`
	CounterServerAddress string `yaml:"counter_server_address"`
	CounterServerPort    string `yaml:"counter_server_port"`
}

type MainServiceHttp struct {
	Port string `yaml:"port"`
}

type CounterReducerConfig struct {
	Kafka CounterReducerKafka `yaml:"kafka"`
	Mongo CounterReducerMongo `yaml:"mongo"`
}
type CounterReducerKafka struct {
	Brokers             []string `yaml:"brokers"`
	CounterReducerTopic string   `yaml:"counter_reducer_topic"`
	CounterReducerGroup string   `yaml:"counter_reducer_group"`
	BatchSize           int      `yaml:"batch_size"`
	Duration            int      `yaml:"duration"`
}

type CounterReducerMongo struct {
	MongoUri        string `yaml:"mongo_uri"`
	MongoDb         string `yaml:"mongo_db"`
	MongoCollection string `yaml:"mongo_collection"`
}




type ScanResultReducerConfig struct {
	Kafka ScanResultReducerKafka `yaml:"kafka"`
	Mongo ScanResultReducerMongo `yaml:"mongo"`
}

type ScanResultReducerKafka struct {
	Brokers []string `yaml:"brokers"`
	ConsumerTopic  string   `yaml:"consumer_topic"`
	ConsumerGroup string   `yaml:"consumer_group"`
	BatchSize int      `yaml:"batch_size"`
	Duration  int      `yaml:"duration"`
}

type ScanResultReducerMongo struct {
	MongoUri string `yaml:"mongo_uri"`
	MongoDb string `yaml:"mongo_db"`
	MongoCollection string `yaml:"mongo_collection"`	
}



type ReportServiceConfig struct {
	Mongo ReportServiceMongo `yaml:"mongo"`
	Grpc  ReportServiceGrpc  `yaml:"grpc"`
	Repository ReportServiceRepository `yaml:"repository"`
}

type ReportServiceMongo struct {
	MongoUri string `yaml:"mongo_uri"`
	MongoDb string `yaml:"mongo_db"`
	MongoCollection string `yaml:"mongo_collection"`
}

type ReportServiceGrpc struct {
	ServerPort string `yaml:"server_port"`
}

type ReportServiceRepository struct {
	Directory string `yaml:"directory"`
}


type MongoCounterSvcConfig struct {
	ScanDirectoriesCount      string
	ScanFilesCount            string
	CompletedDirectoriesCount string
	CompletedFilesCount       string
}
