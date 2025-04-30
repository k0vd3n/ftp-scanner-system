package scanner

import (
	"time"

	"github.com/gabriel-vasile/mimetype"
)

// FileMetaScanner — сканер метаданных (расширения) по magic.
type FileMetaScanner struct{}

func NewFileMetaScanner() *FileMetaScanner {
	return &FileMetaScanner{}
}

// Scan сканирует файл, используя magic, чтобы определить его расширение.
//
// Возвращает строку, содержащую тип файла, например "image/jpeg".
//
// Если возникла ошибка, то возвращает пустую строку и ошибку.
func (s *FileMetaScanner) Scan(filePath string) (string, error) {
	mtype, err := mimetype.DetectFile(filePath)
	if err != nil {
		return "", err
	}
	time.Sleep(150 * time.Millisecond)
	return mtype.String(), nil
}
