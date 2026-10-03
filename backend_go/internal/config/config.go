// 声明环境配置包。
package config

// 开始声明本文件使用的标准库、第三方库和项目内部包。
import (
	// 引入 fmt，用于格式化文本和构造错误。
	"fmt"
	// 引入 net，用于安全拼接主机与端口。
	"net"
	// 引入 os，用于读取环境变量、配置文件或进程信号。
	"os"
	// 引入 strconv，用于转换字符串与整数。
	"strconv"
	// 引入 time，用于处理时间、有效期和超时。
	"time"
	// 引入 time/tzdata，用于嵌入时区数据，让 Windows 等环境也能加载上海时区。
	_ "time/tzdata"

	// 引入 github.com/go-sql-driver/mysql，用于使用 MySQL 驱动、连接配置或错误码。
	driver "github.com/go-sql-driver/mysql"
	// 引入 github.com/joho/godotenv，用于加载 .env 文件中的环境配置。
	"github.com/joho/godotenv"
	// 结束依赖导入列表。
)

// 保存 HTTP、MySQL、Redis、时区和 AI 服务配置。
type Config struct {
	// 保存监听地址以及 MySQL 主机、端口、用户名、密码、库名和字符集。
	Address, DBHost, DBPort, DBUser, DBPassword, DBName, Charset string
	// 保存 Redis 地址、密码和 Go 缓存前缀。
	RedisAddress, RedisPassword, CachePrefix string
	// 保存 Redis 逻辑数据库编号。
	RedisDB int
	// 保存 应用使用的时区。
	Location *time.Location
	// 保存仅在后端使用的 AI 密钥、接口地址和模型名。
	AIKey, AIEndpoint, AIModel string
	// 结束当前作用域的代码块或结构体定义。
}

// 加载环境配置并检查时区和 Redis 数据库编号。
func Load() (Config, error) {
	// 指定了配置文件路径时优先加载该文件。
	if path := os.Getenv("ENV_FILE"); path != "" {
		// 加载指定 .env 文件；解析失败时返回配置错误。
		if err := godotenv.Load(path); err != nil {
			// 配置文件读取失败时返回空配置，并保留原始错误供定位。
			return Config{}, fmt.Errorf("load ENV_FILE: %w", err)
			// 结束当前作用域的代码块或结构体定义。
		}
		// 条件不满足时进入备用处理分支。
	} else {
		// 按目录配置、根目录配置的顺序尝试加载文件。
		for _, path := range []string{".env", "../.env"} {
			// 加载指定 .env 文件；解析失败时返回配置错误。
			if err := godotenv.Load(path); err != nil && !os.IsNotExist(err) {
				// 配置文件格式或读取发生错误时返回空配置和错误原因。
				return Config{}, fmt.Errorf("load configuration: %w", err)
				// 结束当前作用域的代码块或结构体定义。
			}
			// 结束当前作用域的代码块或结构体定义。
		}
		// 结束当前作用域的代码块或结构体定义。
	}
	// 加载应用时区，默认使用上海时区。
	location, err := time.LoadLocation(env("APP_TIMEZONE", "Asia/Shanghai"))
	// 当前操作失败时进入错误处理分支。
	if err != nil {
		// 时区名称无效时返回空配置和时区错误。
		return Config{}, fmt.Errorf("APP_TIMEZONE: %w", err)
		// 结束当前作用域的代码块或结构体定义。
	}
	// 将 Redis 逻辑库编号从环境变量转换为整数。
	redisDB, err := strconv.Atoi(env("REDIS_DB", "0"))
	// Redis 库编号不是整数或为负数时拒绝配置。
	if err != nil || redisDB < 0 {
		// Redis 库编号不合法时返回明确配置错误。
		return Config{}, fmt.Errorf("REDIS_DB must be a non-negative integer")
		// 结束当前作用域的代码块或结构体定义。
	}
	// 组装最终运行配置并返回。
	return Config{
		// 读取当前字段的环境变量，并为未设置的字段使用代码中的默认值。
		Address: env("GO_HTTP_ADDR", "127.0.0.1:8001"),
		// 读取 MySQL 主机和端口，默认连接本机的 3306 端口。
		DBHost: env("DB_HOST", "127.0.0.1"), DBPort: env("DB_PORT", "3306"),
		// 读取 MySQL 用户名和密码；密码不设置默认值。
		DBUser: env("DB_USER", "root"), DBPassword: os.Getenv("DB_PASSWORD"),
		// 读取库名和字符集，并保存已验证的应用时区。
		DBName: env("DB_NAME", "news_app"), Charset: env("DB_CHARSET", "utf8mb4"), Location: location,
		// 读取当前字段的环境变量，并为未设置的字段使用代码中的默认值。
		RedisAddress: net.JoinHostPort(env("REDIS_HOST", "127.0.0.1"), env("REDIS_PORT", "6379")),
		// 保存 Redis 逻辑库、可选密码及独立的 Go 缓存命名空间。
		RedisDB: redisDB, RedisPassword: os.Getenv("REDIS_PASSWORD"), CachePrefix: env("GO_CACHE_PREFIX", "go:news:"),
		// 读取仅供后端使用的 AI 密钥，并选择配置中的模型。
		AIKey: os.Getenv("AI_API_KEY"), AIModel: env("AI_MODEL", "qwen3-max-preview"),
		// 读取当前字段的环境变量，并为未设置的字段使用代码中的默认值。
		AIEndpoint: env("AI_API_ENDPOINT", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"),
		// 返回组装好的配置，并表示加载成功。
	}, nil
	// 结束当前作用域的代码块或结构体定义。
}

// 构造支持特殊密码和时间解析的 MySQL 连接字符串。
func (c Config) DSN() string {
	// 创建 MySQL 驱动的结构化连接配置。
	dsn := driver.NewConfig()
	// 写入用户名、密码和数据库名称。
	dsn.User, dsn.Passwd, dsn.DBName = c.DBUser, c.DBPassword, c.DBName
	// 使用 TCP 连接，并正确拼接主机和端口。
	dsn.Net, dsn.Addr = "tcp", net.JoinHostPort(c.DBHost, c.DBPort)
	// 把数据库日期解析为 Go 时间，并使用应用时区。
	dsn.ParseTime, dsn.Loc = true, c.Location
	// 设置数据库连接字符集，默认支持完整 Unicode。
	dsn.Params = map[string]string{"charset": c.Charset}
	// 设置建连五秒、读写各十秒的超时。
	dsn.Timeout, dsn.ReadTimeout, dsn.WriteTimeout = 5*time.Second, 10*time.Second, 10*time.Second
	// 使用驱动格式化连接字符串，正确处理密码和参数。
	return dsn.FormatDSN()
	// 结束当前作用域的代码块或结构体定义。
}

// 读取环境变量，未设置时使用默认值。
func env(key, fallback string) string {
	// 环境变量有非空值时直接返回该值。
	if value := os.Getenv(key); value != "" {
		// 返回已经取得的配置值或测试数据对象。
		return value
		// 结束当前作用域的代码块或结构体定义。
	}
	// 环境变量未设置时返回默认值。
	return fallback
	// 结束当前作用域的代码块或结构体定义。
}
