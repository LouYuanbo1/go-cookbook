package config

import (
	cryptCfg "github.com/LouYuanbo1/go-webservice/cryptutil/config"
	gormCfg "github.com/LouYuanbo1/go-webservice/gormx/config"
	imgCfg "github.com/LouYuanbo1/go-webservice/imgutil/config"
)

type Config struct {
	DB      gormCfg.DBConfig     `mapstructure:"db"`
	Auth    AuthConfig           `mapstructure:"auth"`
	ImgUtil imgCfg.ImgUtilConfig `mapstructure:"img_util"`
}

type AuthConfig struct {
	CryptUtil cryptCfg.CryptUtilConfig `mapstructure:"crypt_util"`
	Password  string                   `mapstructure:"password"`
	SecretKey string                   `mapstructure:"secret_key"`
	TokenExpire int64                    `mapstructure:"token_expire"`
}
