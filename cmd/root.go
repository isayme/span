package cmd

import (
	"crypto/subtle"
	"fmt"
	"os"

	"github.com/isayme/go-logger"
	"github.com/isayme/span/span"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

		logger.SetFormat(logger.FORMAT_CONSOLE)
		logger.SetLevel(logLevel)

		conf := span.GetConfig()

		upstreamWebdav := conf.Upstream.Webdav
		webdavClient := gowebdav.NewClient(upstreamWebdav.Url, upstreamWebdav.User, upstreamWebdav.Password)
		err = webdavClient.Connect()
		if err != nil {
			logger.Panic(err)
		}

		fs := span.NewFileSystem(webdavClient)
		addr := fmt.Sprintf(":%d", listenPort)
		logger.Infof("服务已启动, 端口: %d ", listenPort)

		app := echo.New()
		app.Use(middleware.RequestID())
		// app.Use(middleware.Logger())

		// basic auth
		webdavConfig := conf.Webdav
		if webdavConfig.User != "" && webdavConfig.Password != "" {
			app.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
				if subtle.ConstantTimeCompare([]byte(username), []byte(webdavConfig.User)) == 1 &&
					subtle.ConstantTimeCompare([]byte(password), []byte(webdavConfig.Password)) == 1 {
					return true, nil
				}
				return false, nil
			}))
		}

		// webdav route
		webdavHandler := &webdav.Handler{
			FileSystem: fs,
			LockSystem: webdav.NewMemLS(),
			// Logger: func(r *http.Request, err error) {
			// 	logger.Infof("webdav method: %s, url: %v, err: %v", r.Method, r.URL.String(), err)
			// },
		}

		app.Any("/*", echo.WrapHandler(webdavHandler))

		logger.Panic(app.Start(addr))
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Panicf("rootCmd execute fail: %s", err.Error())
		os.Exit(1)
	}
}
