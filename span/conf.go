package span

import "github.com/isayme/go-config"

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
	Webdav WebdavConfig `json:"webdav" yaml:"webdav"`

	Upstream UpstreamConfig `json:"upstream" yaml:"upstream"`
}

var cfg Config

func GetConfig() *Config {
	config.Parse(&cfg)

	return &cfg
}
