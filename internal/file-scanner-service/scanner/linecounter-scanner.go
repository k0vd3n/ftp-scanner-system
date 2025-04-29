package scanner

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"time"
)

// ChunkLineCounterScanner — сканер строк блоками по 128К.
type ChunkLineCounterScanner struct{}

func NewChunkLineCounterScanner() *ChunkLineCounterScanner {
	return &ChunkLineCounterScanner{}
}

// Scan сканирует файл, подсчитывая количество строк блоками по 128К.
//
func (s *ChunkLineCounterScanner) Scan(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	const bufSize = 128 * 1024 // 128 КБ
	buf := make([]byte, bufSize)
	var totalLines int
	var carry bool // был ли незавершённый "\n" из предыдущего блока

	for {
		n, err := f.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			count := bytes.Count(chunk, []byte{'\n'})
			// если предыдущий блок не заканчивался \n, первая строка — продолжение
			if carry && chunk[0] != '\n' {
				// лишняя «завершённая» строка не считается
				count--
			}
			totalLines += count
			// подготовим carry: true, если последний байт != '\n'
			carry = chunk[n-1] != '\n'
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
	}

	time.Sleep(15 * time.Second)
	return strconv.Itoa(totalLines), nil
}
