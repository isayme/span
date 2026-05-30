package cmd

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"strings"

	"span/internal"
	"span/internal/config"
	"span/internal/utils"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	webdavfs "github.com/isayme/afero-webdav"
	"github.com/isayme/go-logger"
	"github.com/isayme/go-uuidv4"
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
			internal.ShowVersion()
			os.Exit(0)
		}

		conf := config.GetConfig()

		upstreamWebdav := conf.Upstream.Webdav
		webdavClient := gowebdav.NewClient(upstreamWebdav.Url, upstreamWebdav.User, upstreamWebdav.Password)
		webdavClient.SetHeader("User-Agent", internal.UserAgent)
		webdavClient.SetInterceptor(func(method string, rq *http.Request) {
			reqId, _ := uuidv4.Generate()
			rq.Header.Set("x-request-id", reqId)

			logger.Debugf("webdav method: %s, url: %v, reqId: %v", method, rq.URL.String(), reqId)
		})
		err = webdavClient.Connect()
		if err != nil {
			logger.Panic(err)
		}

		password := conf.Password
		if password == "" {
			password, err := utils.ReadPassword("请输入密码:")
			if err != nil {
				logger.Panicf("读取密码失败: %v", err)
			}
			if internal.IsPasswordTooWeak(password) {
				logger.Panic("密码太弱")
			}
		}

		err = utils.InitBolt("")
		if err != nil {
			logger.Panicf("初始化Bolt失败: %v", err)
		}

		var masterKey []byte
		salt, encryptedMasterKey, hashedAuthKey, err := utils.ReadBolt()
		if err != nil {
			logger.Panicf("读Bolt失败: %v", err)
		}

		if len(salt) > 0 && len(encryptedMasterKey) > 0 && len(hashedAuthKey) > 0 {
			logger.Debug("非首次登录")

			encryptKey, authKey := utils.GenEncryptKeyAndAuthKeyFromPassword(password, salt)
			if subtle.ConstantTimeCompare(utils.HashAuthKey(authKey), hashedAuthKey) != 1 {
				logger.Panic("密码不匹配")
			}

			// authorized, decrypt master key
			masterKey = utils.MustDecryptMasterKey(encryptKey, encryptedMasterKey)
		} else if len(salt) == 0 && len(encryptedMasterKey) == 0 && len(hashedAuthKey) == 0 {
			logger.Debug("首次登录")

			salt = utils.MustRandomSalt()
			masterKey = utils.MustRandomMasterKey()
			encryptKey, authKey := utils.GenEncryptKeyAndAuthKeyFromPassword(password, salt)
			encryptedMasterKey = utils.MustEncryptMasterKey(encryptKey, masterKey)

			utils.WriteBolt(salt, encryptedMasterKey, utils.HashAuthKey(authKey))
		} else {
			logger.Panic("Bolt数据异常")
		}

		fs := internal.NewWebdavFileSystem(internal.NewEncrytFileSystem(webdavfs.New(webdavClient), masterKey))
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

		webdavMethods := []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"HEAD",
			"OPTIONS",
			"PROPFIND",
			"PROPPATCH",
			"MKCOL",
			"COPY",
			"MOVE",
			"LOCK",
			"UNLOCK",
		}

		for _, method := range webdavMethods {
			chi.RegisterMethod(method)
		}
		app := chi.NewRouter()

		webdavConfig := conf.Webdav
		if webdavConfig.User != "" && webdavConfig.Password != "" {
			basicAuthCreds := map[string]string{
				webdavConfig.User: webdavConfig.Password,
			}
			app.Use(middleware.BasicAuth(internal.Name, basicAuthCreds))
		}

		pattern := "/*"
		if webdavConfig.Prefix == "" {
			pattern = fmt.Sprintf("%s/*", strings.TrimRight(webdavConfig.Prefix, "/"))
		}
		for _, method := range webdavMethods {
			app.Method(method, pattern, webdavHandler)
		}

		err = http.ListenAndServe(addr, app)
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
