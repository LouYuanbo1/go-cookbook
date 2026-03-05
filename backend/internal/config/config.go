package config

import (
	"github.com/LouYuanbo1/go-webservice/cryptutil"
	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/LouYuanbo1/go-webservice/imgutil"
)

type Config struct {
	DB      gormx.DBConfig `mapstructure:"db"`
	Auth    AuthConfig     `mapstructure:"auth"`
	ImgUtil imgutil.Config `mapstructure:"img_util"`
}

type AuthConfig struct {
	CryptUtil   cryptutil.Config `mapstructure:"crypt_util"`
	Password    string           `mapstructure:"password"`
	SecretKey   string           `mapstructure:"secret_key"`
	TokenExpire int64            `mapstructure:"token_expire"`
}
