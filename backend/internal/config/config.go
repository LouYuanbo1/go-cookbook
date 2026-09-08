package config

import (
	"go-cookbook/internal/common/utils/img"

	"github.com/LouYuanbo1/go-webservice/cache/driver/redis"
	"github.com/LouYuanbo1/go-webservice/gormx"
)

type Config struct {
	DB    gormx.DBConfig `mapstructure:"db"`
	Redis redis.Config   `mapstructure:"redis"`
	Auth  AuthConfig     `mapstructure:"auth"`
	Img   img.Config     `mapstructure:"img"`
}

type AuthConfig struct {
	Password    string `mapstructure:"password"`
	SecretKey   string `mapstructure:"secret_key"`
	TokenExpire int64  `mapstructure:"token_expire"`
}
