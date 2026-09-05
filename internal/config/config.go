package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	AlertWebhook string `json:",optional"` // 告警通知 webhook URL
}

var (
	Version   string
	BuildDate string
)
