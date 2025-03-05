package scanner

type FileScanner interface {
	Scan(filePath string) (string, error)
}