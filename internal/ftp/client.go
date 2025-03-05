package ftpclient

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
)

type FTPClient struct {
	conn *ftp.ServerConn
}

func NewFTPClient(host, user, pass string) (FtpClientInterface, error) {
	conn, err := ftp.Dial(host, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к FTP: %w", err)
	}

	err = conn.Login(user, pass)
	if err != nil {
		return nil, fmt.Errorf("не удалось войти: %w", err)
	}

	log.Println("✅ Подключение к FTP успешно")
	return &FTPClient{conn: conn}, nil
}

func (c *FTPClient) ListDirectory(path string) ([]string, []string, error) {
	entries, err := c.conn.List(path)
	if err != nil {
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
	return directories, files, nil
}

// Метод скачивания файла
func (c *FTPClient) DownloadFile(remotePath, localDir string) error {
	// Выполняем команду RETR для скачивания файла
	resp, err := c.conn.Retr(remotePath)
	if err != nil {
		return fmt.Errorf(" Ошибка при скачивании %s: %w", remotePath, err)
	}
	defer resp.Close()

	// Определяем локальный путь для сохранения файла
	localPath := filepath.Join(localDir, filepath.Base(remotePath))
	outFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf(" Ошибка при создании файла %s: %w", localPath, err)
	}
	defer outFile.Close()

	// Копируем данные из FTP в локальный файл
	_, err = io.Copy(outFile, resp)
	if err != nil {
		return fmt.Errorf(" Ошибка копирования данных в %s: %w", localPath, err)
	}

	log.Println(" Файл скачан:", localPath)
	return nil
}

func (c *FTPClient) Close() {
	c.conn.Quit()
}
