package span

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path"
	"time"

	"github.com/isayme/go-bufferpool"
	"github.com/isayme/go-logger"
)

type WritableFile struct {
	fs      *FileSystem
	path    string
	wc      io.WriteCloser
	size    int64
	modTime time.Time

	buffer                *bytes.Buffer
	masterKey             []byte
	fileKey               []byte
	encryptFileKeyWritten bool
}

func NewWritableFile(fs *FileSystem, masterKey []byte, path string) *WritableFile {
	return &WritableFile{
		fs:        fs,
		path:      path,
		buffer:    bytes.NewBuffer(nil),
		fileKey:   mustRandomBytes(16),
		masterKey: masterKey,
	}
}

func (file *WritableFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, fmt.Errorf("not support")
}

func (file *WritableFile) Stat() (fs.FileInfo, error) {
	return file, nil
}

func (file *WritableFile) Close() (err error) {
	defer func() {
		if err != nil {
			logger.Warnf("关文件失败, name: %s, err: %v", file.path, err)
		}
	}()

	if file.buffer.Len() > 0 {
		buf := bufferpool.Get(aesBlockSize)
		defer bufferpool.Put(buf)
		n, _ := file.buffer.Read(buf)

		iv := bufferpool.Get(aesBlockSize)
		defer bufferpool.Put(iv)
		genIV(file.size, iv)

		_, err = EncryptFileContent(file.fileKey, iv, buf)
		if err != nil {
			return
		}
		_, err = file.wc.Write(buf[:n])
		if err != nil {
			return
		}
		file.size = file.size + int64(n)
	}

	file.buffer.Reset()

	if file.wc == nil {
		return nil
	}

	return file.wc.Close()
}

func (file *WritableFile) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("not support")
}

func (file *WritableFile) Seek(offset int64, whence int) (int64, error) {
	return 0, fmt.Errorf("not support")
}

func (file *WritableFile) ensureWc() error {
	if file.wc != nil {
		return nil
	}

	rc, wc := io.Pipe()

	go func() {
		file.fs.client.WriteStream(file.fs.resolveName(file.path), rc, FILE_MODE)
	}()
	file.wc = wc
	return nil
}
func (file *WritableFile) writeFileKey() (err error) {
	defer func() {
		if err != nil {
			logger.Warnf("写文件key失败, name: %s, err: %v", file.path, err)
		}
	}()

	if file.encryptFileKeyWritten {
		return nil
	}

	err = file.ensureWc()
	if err != nil {
		return err
	}

	encryptFileKey, err := EncryptFileKey(file.masterKey, file.fileKey)
	if err != nil {
		return err
	}

	_, err = file.wc.Write(encryptFileKey)
	if err != nil {
		return err
	}

	file.encryptFileKeyWritten = true
	return nil
}

func (file *WritableFile) Write(p []byte) (n int, err error) {
	defer func() {
		if err != nil {
			logger.Warnf("写文件失败, name: %s, err: %v", file.path, err)
		}
	}()

	err = file.writeFileKey()
	if err != nil {
		return
	}

	n, err = file.buffer.Write(p)
	if err != nil {
		return
	}

	file.modTime = time.Now()

	iv := bufferpool.Get(aesBlockSize)
	defer bufferpool.Put(iv)

	buf := bufferpool.Get(aesBlockSize)
	defer bufferpool.Put(buf)

	for file.buffer.Len() >= aesBlockSize {
		file.buffer.Read(buf)

		genIV(file.size, iv)

		_, err := EncryptFileContent(file.fileKey, iv, buf)
		if err != nil {
			return 0, err
		}

		_, err = file.wc.Write(buf)
		if err != nil {
			return 0, err
		}
		file.size = file.size + int64(aesBlockSize)
	}

	return
}

func (file *WritableFile) Name() string {
	return path.Base(file.path)
}

func (file *WritableFile) Size() int64 {
	return file.size
}

func (file *WritableFile) Mode() fs.FileMode {
	return FILE_MODE
}
func (file *WritableFile) ModTime() time.Time {
	return file.modTime
}
func (file *WritableFile) IsDir() bool {
	return false
}
func (file *WritableFile) Sys() any {
	return nil
}
