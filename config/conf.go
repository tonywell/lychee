package config

import (
	"os"

	"github.com/spf13/viper"
)

//LogConf 日志配置，用于读取日志配置
type log struct {
	Name       string
	FileName   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Level      string
	Stdout     bool
}

type server struct {
	Address string
	Port    string
}

type TotalConfig struct {
	Log    log
	Server server
}

var config *viper.Viper

var AllConfig TotalConfig

func Init() {
	config = viper.New()
	config.SetConfigName("conf") // 设置文件名称（无后缀）
	config.SetConfigType("yaml") // 设置后缀名 {"1.6以后的版本可以不设置该后缀"}
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	config.AddConfigPath(path + "/config") // 设置文件所在路径
	config.Set("verbose", true)            // 设置默认参数
	//读取配置文件
	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(" Config file not found; ignore error if desired")
		} else {
			panic("Config file was found but another error was produced")
		}
	}
	config.Unmarshal(&AllConfig)
}
