package certprovider

import (
	"errors"
	"strings"
)

type Config struct {
	Vendor    string // "aliyun", "tencent"
	AccessKey string
	SecretKey string
	Email     string
}

func NewProvider(cfg Config) (Provider, error) {
	switch strings.ToLower(cfg.Vendor) {
	case "aliyun":
		return &AliProvider{
			AccessKey: cfg.AccessKey,
			SecretKey: cfg.SecretKey,
			Email:     cfg.Email,
		}, nil
	case "tencent":
		// 返回腾讯云的实现类（结构同上）
		return nil, errors.New("腾讯提供商尚未实施")
	default:
		return nil, errors.New("不支持的供应商")
	}
}
