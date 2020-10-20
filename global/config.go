package global

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
)

type source struct {
	Id           string            `yaml:"id"`
	Input        string            `yaml:"input"`
	Output       string            `yaml:"output"`
	DatabaseType string            `yaml:"database_type"`
	Description  map[string]string `yaml:"description"`
}

type config struct {
	Password   string `yaml:"password"`
	Folder     string `yaml:"folder"`
	IpVersion  int    `yaml:"ip_version"`
	RecordSize int    `yaml:"record_size"`
	Basic      source
	Scene      source
}

var (
	Config config
	Logger *log.Logger
)

func init() {
	viper.SetConfigFile("config.yaml")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig() // 读取配置数据
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	_ = viper.Unmarshal(&Config) // 将配置信息绑定到结构体上

	Logger = log.New(os.Stderr, "[SOCDB] ", log.LstdFlags)
}
