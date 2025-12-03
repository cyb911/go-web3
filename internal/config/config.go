package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort     string
	EthRpcUrl   string
	NetworkName string
	EthPrivate  string

	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

var (
	cfg  *Config
	once sync.Once
)

// Get 获取配置数据，返回值类型
func Get() Config {
	if cfg == nil {
		return MustLoad()
	}
	return *cfg
}

func MustLoad() Config {
	once.Do(func() {
		envPath := findEnvFile()
		if envPath != "" {
			err := godotenv.Load(envPath)
			if err != nil {
				log.Printf("加载 .env 文件失败: %v", err)
			} else {
				log.Printf("加载 .env 文件成功:%s", envPath)
			}
		} else {
			log.Println("未找到 .env 文件，使用系统环境变量")
		}
		cfg = &Config{
			AppPort:     getEnvDefault("APP_PORT", "8080"),
			EthRpcUrl:   os.Getenv("ETH_RPC_URL"),
			NetworkName: os.Getenv("ETH_NETWORK_NAME"),
			EthPrivate:  os.Getenv("ETH_PRIVATE"),

			RedisAddr:     getEnvDefault("REDIS_ADDR", "127.0.0.1:6379"),
			RedisPassword: getEnvDefault("REDIS_PASSWORD", ""),
			RedisDB:       getEnvDefaultInt("REDIS_DB", 0),
		}

		if cfg.EthRpcUrl == "" {
			log.Fatal("配置错误：缺少 ETH_RPC_URL")
		}

		if cfg.NetworkName == "" {
			log.Fatal("配置错误：缺少 ETH_NETWORK_NAME")
		}

		if cfg.EthPrivate == "" {
			log.Fatal("配置错误：缺少 ETH_PRIVATE")
		}
	})
	return *cfg
}

// findEnvFile 从当前目录向上查找 .env 文件
func findEnvFile() string {
	dir, _ := os.Getwd()

	for i := 0; i < 6; i++ {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
		dir = filepath.Dir(dir)
	}
	return ""
}

// 读取环境变脸数据，为空返回默认值
func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvDefaultInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return def // 如果解析失败，也返回默认值
		}
		return int(n)
	}
	return def
}
