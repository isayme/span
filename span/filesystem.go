package span

import (
	"bytes"
	"context"
	"io/fs"
	"os"

	"github.com/isayme/go-logger"
	"github.com/pkg/errors"
	"github.com/studio-b12/gowebdav"
	"golang.org/x/net/webdav"
)

var _ webdav.FileSystem = &FileSystem{}

type FileSystem struct {
	client *gowebdav.Client

	masterKey []byte
}

func NewFileSystem(client *gowebdav.Client, masterKey []byte) webdav.FileSystem {
	return &FileSystem{
		client:    client,
		masterKey: masterKey,
	}
}

func (fs *FileSystem) resolveName(name string) string {
	names := bytes.Split([]byte(name), []byte("/"))

	for idx, item := range names {
		if len(item) > 0 {
			names[idx] = []byte(Base64EncodeToString(MustEncryptFileName(fs.masterKey, item)))
		}
	}

	return string(bytes.Join(names, []byte("/")))
}

func (fs *FileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) (err error) {
	defer func() {
		if err != nil {
			logger.Errorf("新建文件夹 '%s' 失败: %v", name, err)
		} else {
			logger.Infof("新建文件夹 '%s' 成功, perm: %v", name, perm.String())
		}
	}()

	return fs.client.Mkdir(fs.resolveName(name), perm)
}

func (fs *FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (result webdav.File, err error) {
	defer func() {
		if err != nil {
			logger.Errorf("打开文件 '%s' 失败, flag: %x, perm: %s, err: %v", name, flag, perm.String(), err)
		} else {
			logger.Infof("打开文件 '%s' 成功, flag: %x, perm: %s", name, flag, perm.String())
		}
	}()

	if flag&(os.O_SYNC|os.O_APPEND) > 0 {
		return nil, os.ErrInvalid
	}

	if flag&os.O_TRUNC > 0 {
		err := fs.RemoveAll(ctx, name)
		if err != nil {
			if !gowebdav.IsErrNotFound(err) {
				return nil, errors.Wrap(err, "删除源文件失败")
			}
		}
	}

	if flag&os.O_CREATE > 0 {
		return NewWritableFile(fs, fs.masterKey, name), nil
	} else {
		return NewReadableFile(fs, fs.masterKey, name), nil
	}
}

func (fs *FileSystem) RemoveAll(ctx context.Context, name string) (err error) {
	defer func() {
		if err != nil {
			logger.Errorf("删除文件 '%s' 失败: %v", name, err)
		} else {
			logger.Infof("删除文件 '%s' 成功", name)
		}
	}()

	return fs.client.RemoveAll(fs.resolveName(name))
}

func (fs *FileSystem) Rename(ctx context.Context, oldName, newName string) (err error) {
	defer func() {
		if err != nil {
			logger.Errorf("移动文件 '%s' 到 '%s' 失败: %v", oldName, newName, err)
		} else {
			logger.Infof("移动文件 '%s' 到 '%s' 成功", oldName, newName)
		}
	}()

	return fs.client.Rename(fs.resolveName(oldName), fs.resolveName(newName), true)
}

func (fs *FileSystem) Stat(ctx context.Context, name string) (fi os.FileInfo, err error) {
	defer func() {
		if err != nil {
			logger.Errorf("查看文件 '%s' 信息失败: %v", name, err)
		} else {
			logger.Infof("查看文件 '%s' 信息成功, IsDir(): %v, name: %v, mod %v", name, fi.IsDir(), fi.Name(), fi.Mode())
		}
	}()

	fi, err = fs.client.Stat(fs.resolveName(name))
	if err != nil && gowebdav.IsErrNotFound(err) {
		err = os.ErrNotExist
	}

	if err != nil {
		fi = nil
		return
	}

	if name == "/" {
		return
	}

	fi = NewFileInfo(fs.masterKey, fi)
	return
}

func (fs *FileSystem) ReadDir(ctx context.Context, name string) (fis []fs.FileInfo, err error) {
	defer func() {
		if err != nil {
			logger.Errorf("列举文件夹 '%s' 失败: %v", name, err)
		} else {
			logger.Infof("列举文件夹 '%s' 成功, 子文件数: %d", name, len(fis))
		}
	}()

	fis, err = fs.client.ReadDir(fs.resolveName(name))
	for idx := range fis {
		fi := NewFileInfo(fs.masterKey, fis[idx])
		fis[idx] = fi
	}
	return
}
