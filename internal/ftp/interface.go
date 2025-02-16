package ftpclient

type FtpClientInterface interface {
	ListDirectory(path string) ([]string, []string, error)
	Close()
}