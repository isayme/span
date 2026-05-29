package internal

import (
	"io/fs"
	"os"
	"span/internal/constants"
)

const FILE_MODE fs.FileMode = 0600

var _ os.FileInfo = FileInfo{}

type FileInfo struct {
	name string
	os.FileInfo
}

func NewFileInfo(masterKey []byte, fi os.FileInfo) os.FileInfo {
	return FileInfo{
		name:     fi.Name(),
		FileInfo: fi,
	}
}

func (fi FileInfo) Name() string {
	return fi.name
}

func (fi FileInfo) Size() int64 {
	if fi.FileInfo.Size() == 0 {
		return 0
	}
	return fi.FileInfo.Size() - constants.FileKeySize
}
