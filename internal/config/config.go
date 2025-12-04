package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type EthConfig struct {
	RpcUrl      string
	NetworkName string
	Private     string
}

type Config struct {
	appPort     string
	ethConfig   *EthConfig
	redisConfig *RedisConfig
}

var (
	cfg  *Config
	once sync.Once
)

func (c Config) AppPort() string {
	return c.appPort
}

func (c Config) RedisConfig() RedisConfig {
	return *c.redisConfig
}

func (c Config) EthConfig() EthConfig {
	return *c.ethConfig
}

// Get 获取配置数据，返回值类型
func Get() Config {
	if cfg == nil {
		return MustLoad()
	}
	return *cfg
}

func MustLoad() Config {
	once.Do(func() {
		loadEnvFiles()
		cfg = &Config{
			appPort: getEnv("APP_PORT", "8080"),
			ethConfig: &EthConfig{
				RpcUrl:      getEnv("ETH_RPC_URL", ""),
				NetworkName: getEnv("ETH_NETWORK_NAME", ""),
				Private:     getEnv("ETH_PRIVATE", ""),
			},

			redisConfig: &RedisConfig{
				Addr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
				Password: getEnv("REDIS_PASSWORD", ""),
				DB:       getEnv("REDIS_DB", 0),
			},
		}

		validateConfig(cfg)
	})
	return *cfg
}

// findEnvFile 从当前目录向上查找 .env 文件
func findEnvFile() string {
	dir, _ := os.Getwd()

	for i := 0; i < 6; i++ {
		encPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(encPath); err == nil {
			return encPath
		}

		dir = filepath.Dir(dir)
	}
	return ""
}

func loadEnvFiles() {
	envPath := findEnvFile()
	if envPath != "" {
		if err := godotenv.Load(envPath); err != nil {
			log.Printf("加载 .env 文件失败: %v", err)
		} else {
			log.Printf("加载 .env 文件成功: %s", envPath)
		}
	} else {
		log.Println("未找到 .env 文件，使用系统环境变量")
	}
}

// 读取 env 配置数据。def 来在编译期确定类型
func getEnv[T any](key string, def T) T {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	var result any

	switch any(def).(type) {
	case string:
		result = v
	case int:
		n, err := strconv.Atoi(v)
		if err != nil {
			result = def
		} else {
			result = n
		}
	case uint:
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			result = def
		} else {
			result = uint(n)
		}
	case float64:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			result = def
		} else {
			result = f
		}
	case bool:
		b, err := strconv.ParseBool(v)
		if err != nil {
			result = def
		} else {
			result = b
		}
	case time.Duration:
		d, err := time.ParseDuration(v)
		if err != nil {
			result = def
		} else {
			result = d
		}
	default:
		// 不支持的类型
		return def
	}

	return result.(T)

}

func validateConfig(c *Config) {
	ethCfg := c.EthConfig()
	if ethCfg.RpcUrl == "" {
		log.Fatal("配置错误：缺少 ETH_RPC_URL")
	}

	if ethCfg.NetworkName == "" {
		log.Fatal("配置错误：缺少 ETH_NETWORK_NAME")
	}

	if ethCfg.Private == "" {
		log.Fatal("配置错误：缺少 ETH_PRIVATE")
	}
}
