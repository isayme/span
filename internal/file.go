package internal

import (
	"bytes"
	"io"
	"os"
	"span/internal/constants"
	"span/internal/utils"

	"github.com/isayme/go-bufferpool"
	"github.com/spf13/afero"
)

var writtenFlag = 0x01
var readFlag = 0x02

var _ afero.File = &encryptFile{}

type encryptFile struct {
	afero.File

	fileKey []byte

	masterKey []byte

	writeBuffer    *bytes.Buffer
	fileKeyWritten bool
	written        int64

	// read or write, exclusive
	mode int

	// user view of file, not real encrypted file pos
	readBuffer  *bytes.Buffer
	readPos     int64
	fileKeyRead bool
	fileInfo    os.FileInfo
}

func NewEncryptFile(f afero.File, masterKey []byte) afero.File {
	return &encryptFile{
		masterKey:   masterKey,
		File:        f,
		writeBuffer: bytes.NewBuffer(nil),
		readBuffer:  bytes.NewBuffer(nil),
	}
}

func (f *encryptFile) readFileKey() error {
	if f.fileKeyRead {
		return nil
	}

	oldPos := f.readPos

	_, err := f.File.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	encryptedFileKey := bufferpool.Get(constants.FileKeySize)
	defer bufferpool.Put(encryptedFileKey)
	_, err = io.ReadFull(f.File, encryptedFileKey)
	if err != nil {
		return err
	}

	fileKey, err := utils.DecryptFileKey(f.masterKey, encryptedFileKey)
	if err != nil {
		return err
	}
	f.fileKey = fileKey
	f.fileKeyRead = true

	// recovery to old pos
	_, err = f.Seek(oldPos, io.SeekStart)
	if err != nil {
		return err
	}
	return nil
}

func (f *encryptFile) Read(p []byte) (n int, err error) {
	if f.readBuffer.Len() > 0 {
		return f.readBuffer.Read(p)
	}

	err = f.readFileKey()
	if err != nil {
		return
	}

	// Decrypt must operate on AES block boundaries. If readPos is not
	// block-aligned, seek back to the start of the containing block first.
	blockStart := f.readPos / constants.AesBlockSize * constants.AesBlockSize
	skip := int(f.readPos - blockStart)

	if skip > 0 {
		_, err = f.File.Seek(blockStart+constants.FileKeySize, io.SeekStart)
		if err != nil {
			return
		}
	}

	// Read ahead to avoid repeated Seek on small or non-aligned reads.
	// Each AES block was encrypted with a position-based IV
	// (GenIV(blockPos)), so we must decrypt block-by-block even
	// when reading ahead.
	need := len(p) + skip
	if need < constants.AesBlockSize*4 {
		need = constants.AesBlockSize * 4
	}
	if need%constants.AesBlockSize != 0 {
		need = (need/constants.AesBlockSize + 1) * constants.AesBlockSize
	}

	buf := bufferpool.Get(need)
	defer bufferpool.Put(buf)
	nr, err := io.ReadFull(f.File, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		if err == io.EOF {
			return 0, io.EOF
		}
		return 0, err
	}
	if nr == 0 {
		return 0, io.EOF
	}

	iv := bufferpool.Get(constants.AesBlockSize)
	defer bufferpool.Put(iv)
	for off := 0; off < nr; off += constants.AesBlockSize {
		blockEnd := off + constants.AesBlockSize
		if blockEnd > nr {
			blockEnd = nr
		}
		utils.GenIV(blockStart+int64(off), iv)
		_, err = utils.DecryptFileContent(f.fileKey, iv, buf[off:blockEnd])
		if err != nil {
			return
		}
	}

	if skip >= nr {
		return 0, io.EOF
	}

	data := buf[skip:nr]

	f.readPos += int64(len(data))
	n = copy(p, data)
	if remaining := data[n:]; len(remaining) > 0 {
		f.readBuffer.Write(remaining)
	}

	return
}

func (f *encryptFile) Seek(offset int64, whence int) (int64, error) {
	// newOffset := f.File.Seek(offset, whence)
	newOffset := f.readPos
	var err error

	switch whence {
	case io.SeekStart:
		newOffset = offset

		if f.fileKeyRead {
			_, err = f.File.Seek(offset+constants.FileKeySize, whence)
		} else {
			_, err = f.File.Seek(offset, whence)
		}
	case io.SeekCurrent:
		newOffset = newOffset + offset
		_, err = f.File.Seek(offset, whence)
	case io.SeekEnd:
		if f.fileInfo == nil {
			if _, err := f.Stat(); err != nil {
				return 0, err
			}
		}
		newOffset = f.fileInfo.Size() + offset
		_, err = f.File.Seek(offset, whence)
	default:
		return 0, nil
	}

	if newOffset != f.readPos {
		f.readBuffer.Reset()
	}
	f.readPos = newOffset
	return newOffset, err
}

func (f *encryptFile) writeFileKey() error {
	if f.fileKeyWritten {
		return nil
	}

	f.fileKey = utils.MustRandomFileKey()
	encryptedFileKey, err := utils.EncryptFileKey(f.masterKey, f.fileKey)
	if err != nil {
		return err
	}
	_, err = f.File.Write(encryptedFileKey)
	if err != nil {
		return err
	}

	f.fileKeyWritten = true
	return nil
}

func (f *encryptFile) Write(p []byte) (n int, err error) {
	err = f.writeFileKey()
	if err != nil {
		return
	}

	n, err = f.writeBuffer.Write(p)
	if err != nil {
		return
	}

	iv := bufferpool.Get(constants.AesBlockSize)
	defer bufferpool.Put(iv)

	buf := bufferpool.Get(constants.AesBlockSize)
	defer bufferpool.Put(buf)

	for f.writeBuffer.Len() >= constants.AesBlockSize {
		f.writeBuffer.Read(buf)

		utils.GenIV(f.written, iv)

		_, err := utils.EncryptFileContent(f.fileKey, iv, buf)
		if err != nil {
			return 0, err
		}

		_, err = f.File.Write(buf)
		if err != nil {
			return 0, err
		}
		f.written = f.written + int64(constants.AesBlockSize)
	}

	return
}

func (f *encryptFile) Close() error {
	if f.writeBuffer.Len() > 0 {
		buf := bufferpool.Get(constants.AesBlockSize)
		defer bufferpool.Put(buf)
		n, _ := f.writeBuffer.Read(buf)

		iv := bufferpool.Get(constants.AesBlockSize)
		defer bufferpool.Put(iv)
		utils.GenIV(f.written, iv)

		_, err := utils.EncryptFileContent(f.fileKey, iv, buf)
		if err != nil {
			return nil
		}
		_, err = f.File.Write(buf[:n])
		if err != nil {
			return nil
		}
		f.written = f.written + int64(n)
	}

	f.writeBuffer.Reset()
	f.readBuffer.Reset()
	return f.File.Close()
}

func (f *encryptFile) Stat() (os.FileInfo, error) {
	fi, err := f.File.Stat()
	if err != nil {
		return nil, err
	}

	f.fileInfo = NewEncryptFileInfo(fi)
	return f.fileInfo, nil
}

type EncryptFileInfo struct {
	os.FileInfo
}

func NewEncryptFileInfo(fi os.FileInfo) *EncryptFileInfo {
	return &EncryptFileInfo{
		FileInfo: fi,
	}
}

func (fi *EncryptFileInfo) Size() int64 {
	size := fi.FileInfo.Size()
	if size >= constants.FileKeySize {
		return size - constants.FileKeySize
	}

	return 0
}
