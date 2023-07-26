package span

import (
	"io/fs"
	"os"
	"time"
)

const FILE_MODE fs.FileMode = 0600

var _ os.FileInfo = FileInfo{}

type FileInfo struct {
	name string
	fi   os.FileInfo
}

func NewFileInfo(masterKey []byte, fi os.FileInfo) os.FileInfo {
	name, _ := Base64DecodeString(fi.Name())
	name = MustDecryptFileName(masterKey, name)

	return FileInfo{
		name: string(name),
		fi:   fi,
	}
}

func (fi FileInfo) Name() string {
	return fi.name
}

func (fi FileInfo) Size() int64 {
	if fi.fi.Size() == 0 {
		return 0
	}
	return fi.fi.Size() - fileKeySize
}

func (fi FileInfo) Mode() os.FileMode {
	return fi.fi.Mode()
}

func (fi FileInfo) ModTime() time.Time {
	return fi.fi.ModTime()
}

func (fi FileInfo) IsDir() bool {
	return fi.fi.IsDir()
}

func (fi FileInfo) Sys() any {
	return fi.fi.Sys()
}
