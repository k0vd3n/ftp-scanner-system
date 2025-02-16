package ftpclient

import (
	"fmt"
	"log"
	"time"

	"github.com/jlaffaye/ftp"
)

type FTPClient struct {
	conn *ftp.ServerConn
}

func NewFTPClient(host, user, pass string) (*FTPClient, error) {
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

func (c *FTPClient) Close() {
	c.conn.Quit()
}
