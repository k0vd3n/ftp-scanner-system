package config

import (
	"io/fs"
	"os"

	"gopkg.in/yaml.v2"
)

func LoadRoutingConfig(path string) (*RoutingConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg struct {
		Routing RoutingConfig `yaml:"routing"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg.Routing, nil
}

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
	CompletedFilesCountTopic string
	PathForDownloadedFiles   string
	ScannerTypesMap          []string
	// AllTopics                SizeBasedRouterTopics
	Routing RoutingConfig // Добавляем конфиг маршрутизации
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

// type SizeBasedRouterTopics struct {
// 	AllScanResultsTopic string
// 	EmptyResultTopic    string
// 	SmallResultTopic    string
// 	MediumResultTopic   string
// 	LargeResultTopic    string
// }

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
