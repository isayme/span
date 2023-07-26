package span

import (
	"sync"

	"github.com/isayme/go-config"
	"github.com/isayme/go-logger"
)

type Logger struct {
	Format logger.LogFormat `json:"format" yaml:"format"`
	Level  string           `json:"level" yaml:"level"`
}

type WebdavConfig struct {
	Prefix   string `json:"prefix" yaml:"prefix"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
}

type UpstreamWebdavConfig struct {
	Url      string `json:"url" yaml:"url"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
}

type UpstreamConfig struct {
	Webdav UpstreamWebdavConfig `json:"webdav" yaml:"webdav"`
}

type Config struct {
	Logger Logger `json:"logger" yaml:"logger"`

	Password string `json:"password" yaml:"password"`

	Webdav WebdavConfig `json:"webdav" yaml:"webdav"`

	Upstream UpstreamConfig `json:"upstream" yaml:"upstream"`
}

var cfg Config
var loggerInitOnce sync.Once

func GetConfig() *Config {
	config.Parse(&cfg)

	loggerInitOnce.Do(func() {
		logger.SetFormat(cfg.Logger.Format)
		logger.SetLevel(cfg.Logger.Level)
	})

	return &cfg
}
