package cmd

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"github.com/isayme/go-logger"
	"github.com/isayme/span/span"
	"github.com/spf13/cobra"
	"github.com/studio-b12/gowebdav"
	"golang.org/x/net/webdav"
)

var showVersion bool
var listenPort uint16
var logLevel string

func init() {
	rootCmd.Flags().Uint16VarP(&listenPort, "port", "p", 8080, "listen port")
	rootCmd.Flags().StringVarP(&logLevel, "level", "l", "info", "log level")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show version")
}

var rootCmd = &cobra.Command{
	Use: "span",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		if showVersion {
			span.ShowVersion()
			os.Exit(0)
		}

		conf := span.GetConfig()

		upstreamWebdav := conf.Upstream.Webdav
		webdavClient := gowebdav.NewClient(upstreamWebdav.Url, upstreamWebdav.User, upstreamWebdav.Password)
		err = webdavClient.Connect()
		if err != nil {
			logger.Panic(err)
		}

		password := conf.Password
		if password == "" {
			password, err := span.ReadPassword("请输入密码:")
			if err != nil {
				logger.Panicf("读取密码失败: %v", err)
			}
			if span.IsPasswordTooWeak(password) {
				logger.Panic("密码太弱")
			}
		}

		err = span.InitBolt("")
		if err != nil {
			logger.Panicf("初始化Bolt失败: %v", err)
		}

		var masterKey []byte
		salt, encryptMasterKey, authKey, err := span.ReadBolt()
		if err != nil {
			logger.Panicf("读Bolt失败: %v", err)
		}

		if len(salt) > 0 && len(encryptMasterKey) > 0 && len(authKey) > 0 {
			logger.Debug("非首次登录")

			encryptKey, expectAuthKey := span.GenEncryptKeyAndAuthKeyFromPassword(password, salt)
			if hex.EncodeToString(authKey) != hex.EncodeToString(expectAuthKey) {
				logger.Panic("密码不匹配")
			}

			masterKey = span.MustDecryptMasterKey(encryptKey, encryptMasterKey)
		} else if len(salt) == 0 && len(encryptMasterKey) == 0 && len(authKey) == 0 {
			logger.Debug("首次登录")

			salt = span.MustRandomSalt()
			masterKey = span.MustRandomMasterKey()
			encryptKey, authKey := span.GenEncryptKeyAndAuthKeyFromPassword(password, salt)
			encryptMasterKey = span.MustEncryptMasterKey(encryptKey, masterKey)

			span.WriteBolt(salt, encryptMasterKey, authKey)
		} else {
			logger.Panic("Bolt数据异常")
		}

		fs := span.NewFileSystem(webdavClient, masterKey)
		addr := fmt.Sprintf(":%d", listenPort)
		logger.Infof("服务已启动, 端口: %d ", listenPort)

		// webdav route
		webdavHandler := &webdav.Handler{
			FileSystem: fs,
			LockSystem: webdav.NewMemLS(),
			Logger: func(r *http.Request, err error) {
				logger.Infof("webdav method: %s, url: %v, err: %v", r.Method, r.URL.String(), err)
			},
		}

		err = http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// basic auth
			webdavConfig := conf.Webdav
			if webdavConfig.User != "" && webdavConfig.Password != "" {
				username, password, ok := r.BasicAuth()
				if !ok || webdavConfig.User != username || webdavConfig.Password != password {
					w.WriteHeader(401)
					w.Write([]byte("账号密码不匹配"))
					return
				}
			}

			webdavHandler.ServeHTTP(w, r)
		}))
		if err != nil {
			logger.Errorf("启动服务失败: %v", err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Panicf("rootCmd execute fail: %s", err.Error())
		os.Exit(1)
	}
}
