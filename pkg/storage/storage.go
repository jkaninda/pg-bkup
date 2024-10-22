package storage

type Storage interface {
	Copy(fileName string) error
	CopyFrom(fileName string) error
	Prune(retentionDays int) error
	Name() string
}
type Backend struct {
	//Local Path
	LocalPath string
	//Remote path or Destination path
	RemotePath string
}
