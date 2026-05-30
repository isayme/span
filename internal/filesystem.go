package internal

import (
	"context"
	"os"
	"time"

	"github.com/isayme/go-logger"
	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

// EncryptFileSystem implements afero.Fs, delegating to an underlying afero.Fs.
type EncryptFileSystem struct {
	fs        afero.Fs
	masterKey []byte
}

var _ afero.Fs = &EncryptFileSystem{}

func NewEncrytFileSystem(fs afero.Fs, masterKey []byte) *EncryptFileSystem {
	return &EncryptFileSystem{
		fs:        fs,
		masterKey: masterKey,
	}
}

// --- afero.Fs (all methods) ---

func (fs *EncryptFileSystem) Create(name string) (result afero.File, err error) {
	defer func() { logAferoOp("Create", name, err) }()
	file, err := fs.fs.Create(name)
	if err != nil {
		return nil, err
	}
	return NewEncryptFile(file, fs.masterKey), nil
}

func (fs *EncryptFileSystem) Mkdir(name string, perm os.FileMode) (err error) {
	defer func() { logAferoOp("Mkdir", name, err) }()
	return fs.fs.Mkdir(name, perm)
}

func (fs *EncryptFileSystem) MkdirAll(path string, perm os.FileMode) (err error) {
	defer func() { logAferoOp("MkdirAll", path, err) }()
	return fs.fs.MkdirAll(path, perm)
}

func (fs *EncryptFileSystem) Open(name string) (result afero.File, err error) {
	defer func() { logAferoOp("Open", name, err) }()
	file, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return NewEncryptFile(file, fs.masterKey), nil
}

func (fs *EncryptFileSystem) OpenFile(name string, flag int, perm os.FileMode) (result afero.File, err error) {
	defer func() { logAferoOp("OpenFile", name, err) }()
	file, err := fs.fs.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return NewEncryptFile(file, fs.masterKey), nil
}

func (fs *EncryptFileSystem) Remove(name string) (err error) {
	defer func() { logAferoOp("Remove", name, err) }()
	return fs.fs.Remove(name)
}

func (fs *EncryptFileSystem) RemoveAll(path string) (err error) {
	defer func() { logAferoOp("RemoveAll", path, err) }()
	return fs.fs.RemoveAll(path)
}

func (fs *EncryptFileSystem) Rename(oldname, newname string) (err error) {
	defer func() { logAferoOp("Rename", oldname, err) }()
	return fs.fs.Rename(oldname, newname)
}

func (fs *EncryptFileSystem) Stat(name string) (result os.FileInfo, err error) {
	defer func() { logAferoOp("Stat", name, err) }()
	fi, err := fs.fs.Stat(name)
	if err != nil {
		return nil, err
	}

	return NewEncryptFileInfo(fi), nil
	// return fs.fs.Stat(name)
}

func (fs *EncryptFileSystem) Name() string {
	return "span"
}

func (fs *EncryptFileSystem) Chmod(name string, mode os.FileMode) (err error) {
	defer func() { logAferoOp("Chmod", name, err) }()
	return fs.fs.Chmod(name, mode)
}

func (fs *EncryptFileSystem) Chown(name string, uid, gid int) (err error) {
	defer func() { logAferoOp("Chown", name, err) }()
	return fs.fs.Chown(name, uid, gid)
}

func (fs *EncryptFileSystem) Chtimes(name string, atime time.Time, mtime time.Time) (err error) {
	defer func() { logAferoOp("Chtimes", name, err) }()
	return fs.fs.Chtimes(name, atime, mtime)
}

// --- helpers ---

func logAferoOp(op, name string, err error) {
	if err != nil {
		logger.Errorf("%s '%s' 失败: %v", op, name, err)
	} else {
		logger.Infof("%s '%s' 成功", op, name)
	}
}

// runWithContext runs fn in a goroutine and returns its result, or ctx.Err()
// if the context is cancelled before fn completes.
func runWithContext[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	type result struct {
		val T
		err error
	}

	ch := make(chan result, 1)
	go func() {
		var r result
		r.val, r.err = fn()
		ch <- r
	}()

	select {
	case r := <-ch:
		return r.val, r.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

// webdavFS implements webdav.FileSystem, delegating to an underlying afero.Fs
// with context-aware logging. This is a separate type because afero.Fs and
// webdav.FileSystem have conflicting method signatures (Mkdir, OpenFile, etc.).
type webdavFS struct {
	fs afero.Fs
}

var _ webdav.FileSystem = &webdavFS{}

// NewWebdavFileSystem creates a webdav.FileSystem that wraps an afero.Fs.
func NewWebdavFileSystem(fs afero.Fs) webdav.FileSystem {
	return &webdavFS{
		fs: fs,
	}
}

// --- webdav.FileSystem ---

func (w *webdavFS) Mkdir(ctx context.Context, name string, perm os.FileMode) (err error) {
	_, err = runWithContext(ctx, func() (struct{}, error) {
		return struct{}{}, w.fs.Mkdir(name, perm)
	})
	return
}

func (w *webdavFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (result webdav.File, err error) {
	// webdav.Handler uses O_RDWR for PUT, but afero-webdav only supports
	// O_WRONLY for writes. Convert when write flags are present.
	if flag&os.O_RDWR != 0 {
		flag = (flag &^ os.O_RDWR) | os.O_WRONLY
	}

	// afero-webdav's write stream is asynchronous: data is buffered via io.Pipe
	// and only sent to the upstream on Close(). handlePut calls f.Stat() between
	// Copy and Close, which fails if the upstream file doesn't exist yet.
	// Pre-create the file via Create() (which writes an empty file synchronously)
	// so that the subsequent Stat() succeeds.
	if flag&(os.O_WRONLY|os.O_CREATE) != 0 {
		return runWithContext(ctx, func() (webdav.File, error) {
			return w.fs.Create(name)
		})
	}

	return runWithContext(ctx, func() (webdav.File, error) {
		return w.fs.OpenFile(name, flag, perm)
	})
}

func (w *webdavFS) RemoveAll(ctx context.Context, name string) (err error) {
	_, err = runWithContext(ctx, func() (struct{}, error) {
		return struct{}{}, w.fs.RemoveAll(name)
	})
	return
}

func (w *webdavFS) Rename(ctx context.Context, oldName, newName string) (err error) {
	_, err = runWithContext(ctx, func() (struct{}, error) {
		return struct{}{}, w.fs.Rename(oldName, newName)
	})
	return
}

func (w *webdavFS) Stat(ctx context.Context, name string) (fi os.FileInfo, err error) {
	return runWithContext(ctx, func() (os.FileInfo, error) {
		return w.fs.Stat(name)
	})
}
