package e2e

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"span/internal"
	"span/internal/utils"
	"testing"

	"github.com/isayme/go-logger"
	"github.com/stretchr/testify/require"
	"github.com/studio-b12/gowebdav"
	"golang.org/x/net/webdav"
)

func TestE2E(t *testing.T) {
	logger.SetLevel("debug")

	tmpDir := t.TempDir()
	upstreamDir := filepath.Join(tmpDir, "upstream")
	spanDir := filepath.Join(tmpDir, "span")
	require.NoError(t, os.MkdirAll(upstreamDir, 0755))
	require.NoError(t, os.MkdirAll(spanDir, 0755))

	// --- upstream server ---
	upstreamHandler := &webdav.Handler{
		FileSystem: webdav.Dir(upstreamDir), // Folder to share
		LockSystem: webdav.NewMemLS(),
	}

	upstreamMux := http.NewServeMux()
	upstreamMux.Handle("/", upstreamHandler)
	upstreamListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	upstreamServer := &http.Server{Handler: upstreamMux}
	upstreamDone := make(chan error, 1)
	go func() {
		upstreamDone <- upstreamServer.Serve(upstreamListener)
	}()
	defer func() {
		upstreamServer.Shutdown(context.Background())
		<-upstreamDone
	}()
	upstreamAddr := upstreamListener.Addr().String()

	// --- span server ---
	password := "this-is-a-strong-password-that-passes-zxcvbn-1234567890!" + randomString(10)

	boltPath := filepath.Join(spanDir, "intertnal.db")
	require.NoError(t, utils.InitBolt(boltPath))

	salt := utils.MustRandomSalt()
	masterKey := utils.MustRandomMasterKey()
	encryptKey, authKey := utils.GenEncryptKeyAndAuthKeyFromPassword(password, salt)
	encryptedMasterKey := utils.MustEncryptMasterKey(encryptKey, masterKey)
	require.NoError(t, utils.WriteBolt(salt, encryptedMasterKey, utils.HashAuthKey(authKey)))

	// Verify auth (simulate second login)
	readSalt, readEMK, readHAK, err := utils.ReadBolt()
	require.NoError(t, err)
	reEncryptKey, reAuthKey := utils.GenEncryptKeyAndAuthKeyFromPassword(password, readSalt)
	require.Equal(t, utils.HashAuthKey(reAuthKey), readHAK)
	reMasterKey := utils.MustDecryptMasterKey(reEncryptKey, readEMK)
	require.Equal(t, masterKey, reMasterKey)

	upstreamClient := gowebdav.NewClient("http://"+upstreamAddr, "", "")
	upstreamClient.SetHeader("User-Agent", internal.UserAgent)
	require.NoError(t, upstreamClient.Connect())

	fs := internal.NewFileSystem(upstreamClient, masterKey)
	spanHandler := &webdav.Handler{
		FileSystem: fs,
		LockSystem: webdav.NewMemLS(),
	}

	spanMux := http.NewServeMux()
	spanMux.Handle("/", spanHandler)
	spanListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	spanServer := &http.Server{Handler: spanMux}
	spanDone := make(chan error, 1)
	go func() {
		spanDone <- spanServer.Serve(spanListener)
	}()
	defer func() {
		spanServer.Shutdown(context.Background())
		<-spanDone
	}()
	spanAddr := spanListener.Addr().String()

	client := gowebdav.NewClient("http://"+spanAddr, "", "")
	client.SetHeader("User-Agent", "e2e-test")
	require.NoError(t, client.Connect())

	// --- test cases ---
	testCases := []*struct {
		path    string
		size    int
		content string
	}{
		{
			path: "/test1.txt",
			size: 1,
		},

		{
			path: "/test2.txt",
			size: 2,
		},

		{
			path: "/test15.txt",
			size: 15,
		},

		{
			path: "/test16.txt",
			size: 16,
		},

		{
			path: "/test17.txt",
			size: 17,
		},

		{
			path: "/test31.txt",
			size: 31,
		},

		{
			path: "/test32.txt",
			size: 32,
		},

		{
			path: "/test33.txt",
			size: 33,
		},

		{
			path: "/test-1024.txt",
			size: 1024,
		},

		{
			path: "/test-1048576.txt",
			size: 1048576,
		},

		{
			path: "/abc/test33.txt",
			size: 33,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("create folder: %s", tc.path), func(t *testing.T) {
			if "/" != path.Dir(tc.path) {
				require.NoError(t, client.MkdirAll(path.Dir(tc.path), 0755))
			}
		})

		t.Run(fmt.Sprintf("create file: %s", tc.path), func(t *testing.T) {
			tc.content = randomString(tc.size)
			err := client.Write(tc.path, []byte(tc.content), 0644)
			require.NoError(t, err)

			files, err := client.ReadDir(path.Dir(tc.path))
			require.NoError(t, err)
			require.Len(t, files, 1)
			require.Equal(t, path.Base(tc.path), files[0].Name())
			require.EqualValues(t, tc.size, files[0].Size())
		})

		t.Run(fmt.Sprintf("read file: %s", tc.path), func(t *testing.T) {
			data, err := client.Read(tc.path)
			require.NoError(t, err)
			require.Equal(t, tc.content, string(data))
		})

		t.Run(fmt.Sprintf("reverse file: %s", tc.path), func(t *testing.T) {
			newContent := reverse(tc.content)
			tc.content = newContent
			err := client.Write(tc.path, []byte(newContent), 0644)
			require.NoError(t, err)

			data, err := client.Read(tc.path)
			require.NoError(t, err)
			require.Equal(t, string(newContent), string(data))
		})

		t.Run(fmt.Sprintf("write file: %s", tc.path), func(t *testing.T) {
			newContent := randomString(tc.size * 2)
			tc.size = tc.size * 2
			tc.content = newContent
			err := client.Write(tc.path, []byte(newContent), 0644)
			require.NoError(t, err)
			data, err := client.Read(tc.path)
			require.NoError(t, err)
			require.Equal(t, string(newContent), string(data))
		})

		t.Run(fmt.Sprintf("delete file: %s", tc.path), func(t *testing.T) {
			err := client.RemoveAll(tc.path)
			require.NoError(t, err)

			_, err = client.Read(tc.path)
			require.Error(t, err)
		})

		t.Run(fmt.Sprintf("remove dir: %s %v %v", tc.path, "/" != path.Dir(tc.path), path.Dir(tc.path)), func(t *testing.T) {
			if "/" != path.Dir(tc.path) {
				err := client.RemoveAll(path.Dir(tc.path))
				require.NoError(t, err)
			}
		})
	}
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		// Use IntN to pick a random index from the charset
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}

func reverse(s string) string {
	// Convert string to a slice of runes (Unicode code points)
	runes := []rune(s)

	// Swap runes from both ends toward the center
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	// Convert the rune slice back into a string
	return string(runes)
}
