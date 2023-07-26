package span

import (
	"bytes"
	"context"
	"crypto/aes"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"

	"github.com/isayme/go-bufferpool"
	"github.com/isayme/go-logger"
)

type ReadableFile struct {
	fs   *FileSystem
	path string
	rc   io.ReadCloser
	pos  int64

	fileKey   []byte
	masterKey []byte
	buffer    *bytes.Buffer
}

func NewReadableFile(fs *FileSystem, masterKey []byte, path string) *ReadableFile {
	return &ReadableFile{
		fs:        fs,
		path:      path,
		masterKey: masterKey,
		buffer:    bytes.NewBuffer(nil),
	}
}

func (file *ReadableFile) Readdir(count int) (fis []fs.FileInfo, err error) {
	return file.fs.ReadDir(context.Background(), file.path)
}

func (file *ReadableFile) Stat() (fi fs.FileInfo, err error) {
	return file.fs.Stat(context.Background(), file.path)
}

func (file *ReadableFile) Close() error {
	file.buffer.Reset()

	if file.rc == nil {
		return nil
	}

	err := file.rc.Close()
	file.rc = nil
	return err
}

func (file *ReadableFile) readFileKey() (err error) {
	defer func() {
		if err != nil {
			logger.Warnf("读文件密钥失败, name: %s, err: %v", file.path, err)
		}
	}()

	if file.fileKey != nil {
		return nil
	}

	rc, err := file.fs.client.ReadStreamRange(file.fs.resolveName(file.path), 0, fileKeySize)
	if err != nil {
		return err
	}
	defer rc.Close()

	encryptFileKey := bufferpool.Get(aes.BlockSize)
	defer bufferpool.Put(encryptFileKey)
	_, err = io.ReadFull(rc, encryptFileKey)
	if err != nil {
		return err
	}

	fileKey, err := DecryptFileKey(file.masterKey, encryptFileKey)
	if err != nil {
		return err
	}
	file.fileKey = fileKey
	return nil
}

func (file *ReadableFile) ensureRc() (err error) {
	if file.rc != nil {
		return
	}

	file.rc, err = file.fs.client.ReadStreamRange(file.fs.resolveName(file.path), file.pos+fileKeySize, 0)
	return err
}

func (file *ReadableFile) Read(p []byte) (n int, err error) {
	defer func() {
		if err != nil {
			logger.Warnf("读文件失败, name: %s, err: %v", file.path, err)
		}
	}()

	err = file.readFileKey()
	if err != nil {
		return
	}

	err = file.ensureRc()
	if err != nil {
		return
	}

	if file.buffer.Len() > 0 {
		return file.buffer.Read(p)
	}

	buf := bufferpool.Get(aes.BlockSize)
	defer bufferpool.Put(buf)
	nr, err := io.ReadFull(file.rc, buf)
	if err == io.EOF {
		return
	}

	iv := bufferpool.Get(aes.BlockSize)
	defer bufferpool.Put(iv)
	getIv(file.pos, iv)
	_, err = DecryptFileContent(file.fileKey, iv, buf[:nr])
	if err != nil {
		return
	}

	file.pos = file.pos + int64(nr)
	file.buffer.Write(buf[:nr])

	return file.buffer.Read(p)
}

func getIv(pos int64, iv []byte) {
	for i := 0; i < len(iv); i++ {
		iv[i] = 0
	}

	n := pos / int64(aseBlockSize) * int64(aseBlockSize)
	binary.BigEndian.PutUint64(iv[8:], uint64(n))
}

func (file *ReadableFile) Seek(offset int64, whence int) (n int64, err error) {
	err = file.Close()
	if err != nil {
		return 0, err
	}

	pos := file.pos

	switch whence {
	case io.SeekStart:
		pos = offset
	case io.SeekCurrent:
		pos += offset
	case io.SeekEnd:
		s, err := file.Stat()
		if err != nil {
			return 0, err
		}

		pos = s.Size() + offset
	default:
		return 0, fmt.Errorf("not support")
	}

	file.pos = pos
	return pos, nil
}

func (file *ReadableFile) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("not support")
}
