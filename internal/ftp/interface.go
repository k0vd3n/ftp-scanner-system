package ftpclient

type FtpClientInterface interface {
	DownloadFile(remotePath, localDir string) error
	ListDirectory(path string) ([]string, []string, error)
	Close()
}