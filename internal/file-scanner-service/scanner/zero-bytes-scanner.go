package scanner

import (
	"os"
	"strconv"
)

type ZeroBytesScanner struct{}

func NewZeroBytesScanner() *ZeroBytesScanner {
	return &ZeroBytesScanner{}
}

func (s *ZeroBytesScanner) Scan(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, 4096) // Читаем чанками по 4КБ
	zeroCount := 0

	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}

		for i := 0; i < n; i++ {
			if buffer[i] == 0 {
				zeroCount++
			}
		}
	}

	return strconv.Itoa(zeroCount), nil
}
