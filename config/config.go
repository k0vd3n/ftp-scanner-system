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
	FileScanner           FileScannerConfig           `yaml:"file_scanner_service"`
	DirectoryLister       DirectoryListerConfig       `yaml:"directory_lister_service"`
	MainService           MainServiceConfig           `yaml:"main_service"`
	CounterReducer        CounterReducerConfig        `yaml:"counter_reducer_service"`
	ScanResultReducer     ScanResultReducerConfig     `yaml:"scan_result_reducer_service"`
	GenerateReportService GenerateReportServiceConfig `yaml:"generate_report_service"`
	GetReportService      GetReportServiceConfig      `yaml:"get_report_service"`
	StatusService         StatusServiceConfig         `yaml:"status_service"`
	PushGateway           PushGatewayConfig           `yaml:"push_gateway"`
}

type FileScannerConfig struct {
	KafkaConsumer                    FileScannerConsumer                    `yaml:"kafka_consumer"`
	KafkaScanResultProducer          FileScannerProducerScanResults         `yaml:"kafka_scan_result_producer"`
	KafkaCompletedFilesCountProducer FileScannerProducerCompletedFilesCount `yaml:"kafka_completed_files_count_producer"`
	Metrics                          FileScannerMetrics                     `yaml:"metrics"`
	MaxRetries                       int                                    `yaml:"max_retries"`
	TimeoutSeconds                   int                                    `yaml:"timeout_seconds"`
	Mongo                            FileScannerMongo                       `yaml:"mongo"`
}

type FileScannerMongo struct {
	MongoUri        string `yaml:"mongo_uri"`
	MongoDb         string `yaml:"mongo_db"`
	MongoCollection string `yaml:"mongo_collection"`
}

type FileScannerMetrics struct {
	PromHttpPort  string            `yaml:"prom_http_port"`
	InstanceLabel string            `yaml:"instance"`
	PushGateway   PushGatewayConfig `yaml:"push_gateway"`
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
	Kafka   DirectoryListerKafkaConfig `yaml:"kafka"`
	Metrics DirectoryListerPrometheus  `yaml:"metrics"`
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
	MaxRetries                     int    `yaml:"max_retries"`
	TimeoutSeconds                 int    `yaml:"timeout_seconds"`
}

type DirectoryListerPrometheus struct {
	PromHttpPort  string `yaml:"prom_http_port"`
	InstanceLabel string `yaml:"instance"`
}

type MainServiceConfig struct {
	Kafka   MainServiceKafka   `yaml:"kafka"`
	GRPC    MainServiceGrpc    `yaml:"grpc"`
	HTTP    MainServiceHttp    `yaml:"http"`
	Metrics MainServiceMetrics `yaml:"metrics"`
}

type MainServiceMetrics struct {
	PromHttpPort  string `yaml:"prom_http_port"`
	InstanceLabel string `yaml:"instance"`
}

type MainServiceKafka struct {
	Broker          string `yaml:"broker"`
	DirectoryTorpic string `yaml:"directory_topic"`
}

type MainServiceGrpc struct {
	GeneateReportServerAddress string `yaml:"generate_report_server_address"`
	GenerateReportServerPort   string `yaml:"generate_report_server_port"`
	GetReportServerAddress     string `yaml:"get_report_server_address"`
	GetReportServerPort        string `yaml:"get_report_server_port"`
	StatusServerAddress        string `yaml:"status_server_address"`
	StatusServerPort           string `yaml:"status_server_port"`
}

type MainServiceHttp struct {
	Port    string `yaml:"port"`
	WebPath string `yaml:"web_path"`
}

type CounterReducerConfig struct {
	Kafka   CounterReducerKafka   `yaml:"kafka"`
	Mongo   CounterReducerMongo   `yaml:"mongo"`
	Metrics CounterReducerMetrics `yaml:"metrics"`
}

type CounterReducerMetrics struct {
	PromHttpPort  string `yaml:"prom_http_port"`
	InstanceLabel string `yaml:"instance"`
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
	Kafka   ScanResultReducerKafka   `yaml:"kafka"`
	Mongo   ScanResultReducerMongo   `yaml:"mongo"`
	Metrics ScanResultReducerMetrics `yaml:"metrics"`
}

type ScanResultReducerMetrics struct {
	PromHttpPort  string `yaml:"prom_http_port"`
	InstanceLabel string `yaml:"instance"`
}

type ScanResultReducerKafka struct {
	Brokers       []string `yaml:"brokers"`
	ConsumerTopic string   `yaml:"consumer_topic"`
	ConsumerGroup string   `yaml:"consumer_group"`
	BatchSize     int      `yaml:"batch_size"`
	Duration      int      `yaml:"duration"`
}

type ScanResultReducerMongo struct {
	MongoUri        string `yaml:"mongo_uri"`
	MongoDb         string `yaml:"mongo_db"`
	MongoCollection string `yaml:"mongo_collection"`
}

type GetReportServiceConfig struct {
	Mongo   GetReportServiceMongo   `yaml:"mongo"`
	Grpc    GetReportServiceGrpc    `yaml:"grpc"`
	Metrics GetReportServiceMetrics `yaml:"metrics"`
}

type GetReportServiceMongo struct {
	MongoUri        string `yaml:"mongo_uri"`
	MongoDb         string `yaml:"mongo_db"`
	MongoCollection string `yaml:"mongo_collection"`
}

type GetReportServiceGrpc struct {
	ServerPort string `yaml:"server_port"`
}

type GetReportServiceMetrics struct {
	PromHttpPort  string `yaml:"prom_http_port"`
	InstanceLabel string `yaml:"instance"`
}

type GenerateReportServiceConfig struct {
	Mongo      GenerateReportServiceMongo      `yaml:"mongo"`
	Grpc       GenerateReportServiceGrpc       `yaml:"grpc"`
	Repository GenerateReportServiceRepository `yaml:"repository"`
	Metrics    GenerateReportServiceMetrics    `yaml:"metrics"`
}

type GenerateReportServiceMetrics struct {
	PromHttpPort  string `yaml:"prom_http_port"`
	InstanceLabel string `yaml:"instance"`
}

type GenerateReportServiceMongo struct {
	MongoUri         string `yaml:"mongo_uri"`
	MongoDb          string `yaml:"mongo_db"`
	MongoCollection1 string `yaml:"mongo_collection1"`
	MongoCollection2 string `yaml:"mongo_collection2"`
}

type GenerateReportServiceGrpc struct {
	ServerPort string `yaml:"server_port"`
}

type GenerateReportServiceRepository struct {
	Directory string `yaml:"directory"`
}

type StatusServiceConfig struct {
	Mongo   StatusServiceMongo   `yaml:"mongo"`
	Grpc    StatusServiceGrpc    `yaml:"grpc"`
	Metrics StatusServiceMetrics `yaml:"metrics"`
}

type StatusServiceMetrics struct {
	PromHttpPort  string `yaml:"prom_http_port"`
	InstanceLabel string `yaml:"instance"`
}

type StatusServiceMongo struct {
	MongoUri                  string `yaml:"mongo_uri"`
	MongoDb                   string `yaml:"mongo_db"`
	ScanDirectoriesCount      string `yaml:"scan_directories_count_collection"`
	ScanFilesCount            string `yaml:"scan_files_count_collection"`
	CompletedDirectoriesCount string `yaml:"completed_directories_count_collection"`
	CompletedFilesCount       string `yaml:"completed_files_count_collection"`
}

type StatusServiceGrpc struct {
	Port string `yaml:"port"`
}

type PushGatewayConfig struct {
	URL          string `yaml:"url"`
	JobName      string `yaml:"job_name"`
	Instance     string `yaml:"instance"`
	PushInterval int    `yaml:"push_interval"`
}

// type JaegerConfig struct {
// 	ServiceName string         `yaml:"service_name"`
// 	AgentHost   string         `yaml:"agent_host"`
// 	AgentPort   string         `yaml:"agent_port"`
// 	Sampler     JaegerSampler  `yaml:"sampler"`
// 	Reporter    JaegerReporter `yaml:"reporter"`
// }

type JaegerSampler struct {
	Type  string  `yaml:"type"`
	Param float64 `yaml:"param"`
}

type JaegerReporter struct {
	LogSpans bool `yaml:"log_spans"`
}
