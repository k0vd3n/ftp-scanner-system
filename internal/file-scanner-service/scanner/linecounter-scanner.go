package scanner

import (
	"bytes"
	"io"
	"os"
	"strconv"
)

// ChunkLineCounterScanner — сканер строк блоками по 128К.
type ChunkLineCounterScanner struct{}

func NewChunkLineCounterScanner() *ChunkLineCounterScanner {
	return &ChunkLineCounterScanner{}
}

// Scan сканирует файл, подсчитывая количество строк блоками по 128К.
func (s *ChunkLineCounterScanner) Scan(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return strconv.Itoa(0), err
	}
	defer f.Close()

	const bufSize = 128 * 1024
	buf := make([]byte, bufSize)
	totalNewlines := 0
	var lastByte byte

	for {
		n, err := f.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			totalNewlines += bytes.Count(chunk, []byte{'\n'})
			lastByte = chunk[n-1]
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return strconv.Itoa(0), err
		}
	}

	// Если файл не пуст и последний байт не '\n', то последняя строка "неоконченная"
	if lastByte != 0 && lastByte != '\n' {
		return strconv.Itoa(totalNewlines + 1), nil
	}
	return strconv.Itoa(totalNewlines), nil
}
