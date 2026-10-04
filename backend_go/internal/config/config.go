package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
	_ "time/tzdata"

	driver "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Config struct {
	Address, DBHost, DBPort, DBUser, DBPassword, DBName, Charset string
	RedisAddress, RedisPassword, CachePrefix                     string
	RedisDB                                                      int
	Location                                                     *time.Location
	AIKey, AIEndpoint, AIModel                                   string
}

// 加载环境配置并检查时区和 Redis 数据库编号。
func Load() (Config, error) {
	if path := os.Getenv("ENV_FILE"); path != "" {
		if err := godotenv.Load(path); err != nil {
			return Config{}, fmt.Errorf("load ENV_FILE: %w", err)
		}
	} else {
		for _, path := range []string{".env", "../.env"} {
			if err := godotenv.Load(path); err != nil && !os.IsNotExist(err) {
				return Config{}, fmt.Errorf("load configuration: %w", err)
			}
		}
	}
	location, err := time.LoadLocation(env("APP_TIMEZONE", "Asia/Shanghai"))
	if err != nil {
		return Config{}, fmt.Errorf("APP_TIMEZONE: %w", err)
	}
	redisDB, err := strconv.Atoi(env("REDIS_DB", "0"))
	if err != nil || redisDB < 0 {
		return Config{}, fmt.Errorf("REDIS_DB must be a non-negative integer")
	}
	return Config{
		Address: env("GO_HTTP_ADDR", "127.0.0.1:8001"),
		DBHost:  env("DB_HOST", "127.0.0.1"), DBPort: env("DB_PORT", "3306"),
		DBUser: env("DB_USER", "root"), DBPassword: os.Getenv("DB_PASSWORD"),
		DBName: env("DB_NAME", "news_app"), Charset: env("DB_CHARSET", "utf8mb4"), Location: location,
		RedisAddress: net.JoinHostPort(env("REDIS_HOST", "127.0.0.1"), env("REDIS_PORT", "6379")),
		RedisDB:      redisDB, RedisPassword: os.Getenv("REDIS_PASSWORD"), CachePrefix: env("GO_CACHE_PREFIX", "go:news:"),
		AIKey: os.Getenv("AI_API_KEY"), AIModel: env("AI_MODEL", "qwen3-max-preview"),
		AIEndpoint: env("AI_API_ENDPOINT", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"),
	}, nil
}

// 构造支持特殊密码和时间解析的 MySQL 连接字符串。
func (c Config) DSN() string {
	dsn := driver.NewConfig()
	dsn.User, dsn.Passwd, dsn.DBName = c.DBUser, c.DBPassword, c.DBName
	dsn.Net, dsn.Addr = "tcp", net.JoinHostPort(c.DBHost, c.DBPort)
	dsn.ParseTime, dsn.Loc = true, c.Location
	dsn.Params = map[string]string{"charset": c.Charset}
	dsn.Timeout, dsn.ReadTimeout, dsn.WriteTimeout = 5*time.Second, 10*time.Second, 10*time.Second
	return dsn.FormatDSN()
}

// 读取环境变量，未设置时使用默认值。
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
