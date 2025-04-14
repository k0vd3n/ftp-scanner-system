package logger

import (
	"go.uber.org/zap"
)

// InitLogger создаёт и настраивает глобальный логгер.
// Используем zap-production (JSON) с выводом в stdout.
func InitLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	// Если логи собираются через ELK, достаточно stdout.
	cfg.OutputPaths = []string{"stdout"}
	// Возможна дополнительная настройка уровня логирования и формата.
	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return logger, nil
}
