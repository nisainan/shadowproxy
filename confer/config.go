package confer

import (
	"fmt"

	"github.com/spf13/viper"
)

var confer *Confer

type Confer struct {
	ListenAddress     string `yaml:"listen-address" mapstructure:"listen-address"`           // 代理监听地址
	Username          string `yaml:"username" mapstructure:"username"`                       // 鉴权用户名
	Password          string `yaml:"password" mapstructure:"password"`                       // 鉴权密码
	ProbeResistDomain string `yaml:"probe-resist-domain" mapstructure:"probe-resist-domain"` // 鉴权域名
	CertFile          string `yaml:"cert-file" mapstructure:"cert-file"`                     // Cert地址
	KeyFile           string `yaml:"key-file" mapstructure:"key-file"`                       // Key地址
	CheatHost         string `yaml:"cheat-host" mapstructure:"cheat-host"`                   // 欺骗的内网地址
}

func InitConfer(configURL string) (err error) {
	v := viper.New()
	v.SetConfigFile(configURL)
	err = v.ReadInConfig()
	if err != nil {
		err = fmt.Errorf("fatal error config file: %w", err)
		return
	}
	if err = v.Unmarshal(&confer); err != nil {
		return err
	}
	return
}

func GlobalConfig() *Confer {
	return confer
}
