package ftpclient

import (
	"fmt"
	filescannerservice "ftp-scanner_try2/internal/file-scanner-service"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
	"go.uber.org/zap"
)

type FTPClient struct {
	conn   *ftp.ServerConn
	logger *zap.Logger
}

func NewFTPClient(host, user, pass string, logger *zap.Logger) (FtpClientInterface, error) {
	logger.Info("FTPClient: Подключение к FTP",
		zap.String("host", host),
		zap.String("user", user))
	conn, err := ftp.Dial(host, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		logger.Error("FTPClient: Не удалось подключиться к FTP", zap.Error(err))
		return nil, fmt.Errorf("не удалось подключиться к FTP: %w", err)
	}

	err = conn.Login(user, pass)
	if err != nil {
		logger.Error("FTPClient: Не удалось войти", zap.Error(err))
		return nil, fmt.Errorf("не удалось войти: %w", err)
	}

	logger.Info("FTPClient: Подключение успешно")
	return &FTPClient{conn: conn, logger: logger}, nil
}

func (c *FTPClient) ListDirectory(path string) ([]string, []string, error) {
	c.logger.Info("FTPClient: Листинг директории", zap.String("path", path))
	entries, err := c.conn.List(path)
	if err != nil {
		c.logger.Error("FTPClient: Ошибка при листинге директории", zap.String("path", path), zap.Error(err))
		return nil, nil, err
	}

	var directories, files []string
	for _, entry := range entries {
		if entry.Name == "." || entry.Name == ".." {
			continue
		}
		fullPath := path + "/" + entry.Name
		if entry.Type == ftp.EntryTypeFolder {
			directories = append(directories, fullPath)
		} else {
			files = append(files, fullPath)
		}
	}
	c.logger.Info("FTPClient: Листинг завершен",
		zap.String("path", path),
		zap.Int("directories", len(directories)),
		zap.Int("files", len(files)))
	return directories, files, nil
}

// Метод скачивания файла
func (c *FTPClient) DownloadFile(remotePath, localDir string) error {
	c.logger.Info("FTPClient: Скачивание файла", zap.String("remotePath", remotePath))
	startTime := time.Now()
	// Выполняем команду RETR для скачивания файла
	resp, err := c.conn.Retr(remotePath)
	if err != nil {
		log.Printf("ftpclient: Ошибка при скачивании %s: %v", remotePath, err)
		return fmt.Errorf(" Ошибка при скачивании %s: %w", remotePath, err)
	}
	defer resp.Close()
	filescannerservice.DownloadDuration.Observe(time.Since(startTime).Seconds())
	filescannerservice.DownloadedFiles.Inc()

	c.logger.Info("FTPClient: Файл скачан", zap.String("remotePath", remotePath))
	startTime = time.Now()
	// Определяем локальный путь для сохранения файла
	localPath := filepath.Join(localDir, filepath.Base(remotePath))
	outFile, err := os.Create(localPath)
	if err != nil {
		c.logger.Error("FTPClient: Ошибка создания файла", zap.String("localPath", localPath), zap.Error(err))
		return fmt.Errorf(" Ошибка при создании файла %s: %w", localPath, err)
	}
	defer outFile.Close()

	// Копируем данные из FTP в локальный файл
	_, err = io.Copy(outFile, resp)
	if err != nil {
		c.logger.Error("FTPClient: Ошибка копирования данных", zap.String("localPath", localPath), zap.Error(err))
		return fmt.Errorf(" Ошибка копирования данных в %s: %w", localPath, err)
	}
	filescannerservice.SavingDuration.Observe(time.Since(startTime).Seconds())
	filescannerservice.SavedFilesCounter.Inc()

	c.logger.Info("FTPClient: Файл успешно скачан и сохранён", zap.String("localPath", localPath))
	return nil
}

func (c *FTPClient) Close() {
	c.logger.Info("FTPClient: Закрытие соединения с FTP")
	c.conn.Quit()
}

func (c *FTPClient) CheckConnection() error {
	// Простая команда для проверки соединения
	return c.conn.NoOp()
}
