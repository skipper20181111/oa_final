package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	AuthBackEnd struct {
		AccessSecret string
		AccessExpire int64
	}
	DB struct {
		DataSource string
	}
	Cache      cache.CacheConf
	WxConf     WxConf
	SfConf     SfConf
	ServerInfo struct {
		Url string
	}
	Lock struct {
		Host string
	}
}
