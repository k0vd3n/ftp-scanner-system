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
	log.Printf("ftpclient: Подключение к FTP ip: %s user: %s...", host, user)
	log.Println("ftpclient: Подключение к FTP...")
	conn, err := ftp.Dial(host, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Println("ftpclient: Не удалось подключиться к FTP:", err)
		return nil, fmt.Errorf("не удалось подключиться к FTP: %w", err)
	}

	err = conn.Login(user, pass)
	if err != nil {
		log.Println("ftpclient: Не удалось войти:", err)
		return nil, fmt.Errorf("не удалось войти: %w", err)
	}

	log.Println("ftpclient: Подключение к FTP успешно")
	return &FTPClient{conn: conn}, nil
}

func (c *FTPClient) ListDirectory(path string) ([]string, []string, error) {
	log.Printf("ftpclient: Просмотр содержимого директории %s...", path)
	entries, err := c.conn.List(path)
	if err != nil {
		log.Printf("ftpclient: Ошибка при просмотре содержимого директории %s: %v", path, err)
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
	log.Printf("ftpclient: Просмотр содержимого директории %s завершен", path)
	log.Printf("ftpclient: Найдено %d поддиректорий и %d файлов", len(directories), len(files))
	return directories, files, nil
}

// Метод скачивания файла
func (c *FTPClient) DownloadFile(remotePath, localDir string) error {
	log.Printf("ftpclient: Скачивание файла %s...", remotePath)
	// Выполняем команду RETR для скачивания файла
	resp, err := c.conn.Retr(remotePath)
	if err != nil {
		log.Printf("ftpclient: Ошибка при скачивании %s: %v", remotePath, err)
		return fmt.Errorf(" Ошибка при скачивании %s: %w", remotePath, err)
	}
	defer resp.Close()

	// Определяем локальный путь для сохранения файла
	localPath := filepath.Join(localDir, filepath.Base(remotePath))
	outFile, err := os.Create(localPath)
	if err != nil {
		log.Printf("ftpclient: Ошибка при создании файла %s: %v", localPath, err)
		return fmt.Errorf(" Ошибка при создании файла %s: %w", localPath, err)
	}
	defer outFile.Close()

	log.Printf("ftpclient: Копирование данных в %s...", localPath)
	// Копируем данные из FTP в локальный файл
	_, err = io.Copy(outFile, resp)
	if err != nil {
		log.Printf("ftpclient: Ошибка копирования данных в %s: %v", localPath, err)
		return fmt.Errorf(" Ошибка копирования данных в %s: %w", localPath, err)
	}

	log.Println("ftpclient: Файл скачан и сохранен в:", localPath)
	return nil
}

func (c *FTPClient) Close() {
	log.Println("ftpclient: Закрытие соединения с FTP...")
	c.conn.Quit()
}
