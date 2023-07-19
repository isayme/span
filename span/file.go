package span

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"path"
	"time"
)

type ReadableFile struct {
	fs   *FileSystem
	path string
	rc   io.ReadCloser
	pos  int64
}

func NewReadableFile(fs *FileSystem, path string) *ReadableFile {
	return &ReadableFile{
		fs:   fs,
		path: path,
	}
}

func (file *ReadableFile) Readdir(count int) (fis []fs.FileInfo, err error) {
	return file.fs.ReadDir(context.Background(), file.path)
}

func (file *ReadableFile) Stat() (fi fs.FileInfo, err error) {
	return file.fs.Stat(context.Background(), file.path)
}

func (file *ReadableFile) Close() error {
	if file.rc == nil {
		return nil
	}

	err := file.rc.Close()
	file.rc = nil
	return err
}

func (file *ReadableFile) ensureRc() (err error) {
	if file.rc != nil {
		return
	}

	file.rc, err = file.fs.client.ReadStreamRange(file.path, file.pos, 0)
	return err
}

func (file *ReadableFile) Read(p []byte) (n int, err error) {
	err = file.ensureRc()
	if err != nil {
		return
	}
	return file.rc.Read(p)
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

type WritableFile struct {
	fs      *FileSystem
	path    string
	wc      io.WriteCloser
	size    int64
	modTime time.Time
}

func NewWritableFile(fs *FileSystem, path string) *WritableFile {
	rc, wc := io.Pipe()

	go func() {
		fs.client.WriteStream(path, rc, 0660)
	}()

	return &WritableFile{
		fs:   fs,
		path: path,
		wc:   wc,
	}
}

func (file *WritableFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, fmt.Errorf("not support")
}

func (file *WritableFile) Stat() (fs.FileInfo, error) {
	return file, nil
}

func (file *WritableFile) Close() error {
	return file.wc.Close()
}

func (file *WritableFile) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("not support")
}

func (file *WritableFile) Seek(offset int64, whence int) (int64, error) {
	return 0, fmt.Errorf("not support")
}

func (file *WritableFile) Write(p []byte) (n int, err error) {
	n, err = file.wc.Write(p)
	file.size = file.size + int64(n)
	file.modTime = time.Now()
	return
}

func (file *WritableFile) Name() string {
	return path.Base(file.path)
}

func (file *WritableFile) Size() int64 {
	return file.size
}

func (file *WritableFile) Mode() fs.FileMode {
	return 0660
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
