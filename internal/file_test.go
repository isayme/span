package internal

import (
	"io"
	"testing"

	"span/internal/utils"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// readAll reads exactly len(buf) bytes from r into buf, looping until all
// bytes are read (like io.ReadFull but returns partial results on error).
func readAll(r io.Reader, buf []byte) (int, error) {
	var total int
	for total < len(buf) {
		n, err := r.Read(buf[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func TestEncryptFileSeekAndRead(t *testing.T) {
	require := require.New(t)
	mem := afero.NewMemMapFs()
	masterKey := utils.MustRandomMasterKey()
	fs := NewEncrytFileSystem(mem, masterKey)

	// 5 full blocks (80 bytes): 8 groups of 10 chars each
	plaintext := make([]byte, 80)
	for i := range plaintext {
		plaintext[i] = 'A' + byte(i/10)
	}

	// --- write ---
	f, err := fs.Create("/test.txt")
	require.NoError(err)
	_, err = f.Write(plaintext)
	require.NoError(err)
	err = f.Close()
	require.NoError(err)

	// --- open for read ---
	f, err = fs.Open("/test.txt")
	require.NoError(err)
	defer f.Close()

	t.Run("non-block-aligned seek within block 0", func(t *testing.T) {
		pos, err := f.Seek(13, io.SeekStart)
		require.NoError(err)
		require.Equal(int64(13), pos)
		buf := make([]byte, 10)
		total, err := readAll(f, buf)
		require.NoError(err)
		require.Equal(10, total)
		require.Equal(plaintext[13:23], buf)
	})

	t.Run("seek to block boundary", func(t *testing.T) {
		pos, err := f.Seek(16, io.SeekStart)
		require.NoError(err)
		require.Equal(int64(16), pos)
		buf := make([]byte, 16)
		total, err := readAll(f, buf)
		require.NoError(err)
		require.Equal(16, total)
		require.Equal(plaintext[16:32], buf)
	})

	t.Run("non-block-aligned seek within block 1", func(t *testing.T) {
		pos, err := f.Seek(20, io.SeekStart)
		require.NoError(err)
		require.Equal(int64(20), pos)
		buf := make([]byte, 10)
		total, err := readAll(f, buf)
		require.NoError(err)
		require.Equal(10, total)
		require.Equal(plaintext[20:30], buf)
	})

	t.Run("read near end returns remaining bytes then EOF", func(t *testing.T) {
		pos, err := f.Seek(64, io.SeekStart)
		require.NoError(err)
		require.Equal(int64(64), pos)
		buf := make([]byte, 16)
		total, err := readAll(f, buf)
		require.NoError(err)
		require.Equal(16, total)
		require.Equal(plaintext[64:], buf)

		// Next read must return EOF
		n, err := f.Read(make([]byte, 1))
		require.Equal(0, n)
		require.ErrorIs(err, io.EOF)
	})

	t.Run("seek after partial read resets readBuffer", func(t *testing.T) {
		pos, err := f.Seek(0, io.SeekStart)
		require.NoError(err)
		require.Equal(int64(0), pos)

		partRead := make([]byte, 10)
		total, err := readAll(f, partRead)
		require.NoError(err)
		require.Equal(10, total)
		require.Equal(plaintext[:10], partRead)

		// seek to a different location while readBuffer still has data
		pos, err = f.Seek(30, io.SeekStart)
		require.NoError(err)
		require.Equal(int64(30), pos)

		buf := make([]byte, 10)
		total, err = readAll(f, buf)
		require.NoError(err)
		require.Equal(10, total)
		require.Equal(plaintext[30:40], buf)
	})

	t.Run("read entire file from start", func(t *testing.T) {
		pos, err := f.Seek(0, io.SeekStart)
		require.NoError(err)
		require.Equal(int64(0), pos)

		result, err := io.ReadAll(f)
		require.NoError(err)
		require.Equal(plaintext, result)
	})
}

func TestEncryptFileSeekEnd(t *testing.T) {
	require := require.New(t)
	mem := afero.NewMemMapFs()
	masterKey := utils.MustRandomMasterKey()
	fs := NewEncrytFileSystem(mem, masterKey)

	plaintext := make([]byte, 80)
	for i := range plaintext {
		plaintext[i] = 'A' + byte(i/10)
	}

	f, err := fs.Create("/test.txt")
	require.NoError(err)
	_, err = f.Write(plaintext)
	require.NoError(err)
	err = f.Close()
	require.NoError(err)

	f, err = fs.Open("/test.txt")
	require.NoError(err)
	defer f.Close()

	// SeekEnd with negative offset = read last N bytes
	pos, err := f.Seek(-10, io.SeekEnd)
	require.NoError(err)
	require.Equal(int64(len(plaintext)-10), pos)

	buf := make([]byte, 10)
	_, err = io.ReadFull(f, buf)
	require.NoError(err)
	require.Equal(plaintext[len(plaintext)-10:], buf)

	// Next read returns EOF
	n, err := f.Read(make([]byte, 1))
	require.Equal(0, n)
	require.ErrorIs(err, io.EOF)

	// SeekCurrent backwards (20 bytes to land on the block before last)
	pos, err = f.Seek(-20, io.SeekCurrent)
	require.NoError(err)
	require.Equal(int64(len(plaintext)-20), pos)

	buf2 := make([]byte, 5)
	n2, err := readAll(f, buf2)
	require.NoError(err)
	require.Equal(5, n2)
	require.Equal(plaintext[len(plaintext)-20:len(plaintext)-15], buf2)

	// SeekCurrent forward
	pos, err = f.Seek(0, io.SeekStart)
	require.NoError(err)
	pos, err = f.Seek(10, io.SeekCurrent)
	require.NoError(err)
	require.Equal(int64(10), pos)

	buf3 := make([]byte, 5)
	n3, err := readAll(f, buf3)
	require.NoError(err)
	require.Equal(5, n3)
	require.Equal(plaintext[10:15], buf3)
}
