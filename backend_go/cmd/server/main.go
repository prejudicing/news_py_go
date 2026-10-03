// 声明可执行程序的入口包。
package main

// 开始声明本文件使用的标准库、第三方库和项目内部包。
import (
	// 引入 context，用于管理请求取消、超时和资源释放。
	"context"
	// 引入 errors，用于判断错误类型与包装后的错误。
	"errors"
	// 引入 log，用于输出服务运行日志。
	"log"
	// 引入 net/http，用于构造 HTTP 请求、客户端和服务。
	"net/http"
	// 引入 os，用于读取环境变量、配置文件或进程信号。
	"os"
	// 引入 os/signal，用于将系统退出信号转换为上下文取消事件。
	"os/signal"
	// 引入 syscall，用于使用系统信号常量。
	"syscall"
	// 引入 time，用于处理时间、有效期和超时。
	"time"

	// 引入 github.com/gin-gonic/gin，用于使用 Gin 路由、请求上下文和 JSON 响应。
	"github.com/gin-gonic/gin"
	// 引入项目内部的 cache 包，复用项目 Redis 缓存封装。
	"github.com/prejudicing/news_py_go/backend_go/internal/cache"
	// 引入项目内部的 config 包，复用项目环境配置。
	"github.com/prejudicing/news_py_go/backend_go/internal/config"
	// 引入项目内部的 server 包，创建项目业务服务。
	"github.com/prejudicing/news_py_go/backend_go/internal/server"
	// 引入 github.com/redis/go-redis/v9，用于使用 Redis 客户端读取和写入缓存。
	"github.com/redis/go-redis/v9"
	// 引入 gorm.io/driver/mysql，用于使用 GORM 的 MySQL 适配器。
	"gorm.io/driver/mysql"
	// 引入 gorm.io/gorm，用于使用 GORM 执行数据库查询和事务。
	"gorm.io/gorm"
	// 引入 gorm.io/gorm/logger，用于配置 GORM 日志级别。
	"gorm.io/gorm/logger"
	// 结束依赖导入列表。
)

// 程序入口，启动服务并处理启动失败。
func main() {
	// 当前操作失败时进入错误处理分支。
	if err := run(); err != nil {
		// 输出致命启动错误并退出进程。
		log.Fatal(err)
		// 结束当前作用域的代码块或结构体定义。
	}
	// 结束当前作用域的代码块或结构体定义。
}

// 初始化依赖，启动 HTTP 服务并处理退出信号。
func run() error {
	// 加载环境配置，并保存可能发生的配置错误。
	cfg, err := config.Load()
	// 当前操作失败时进入错误处理分支。
	if err != nil {
		// 把当前错误返回给调用方；在事务回调中这会触发回滚。
		return err
		// 结束当前作用域的代码块或结构体定义。
	}
	// 连接 MySQL 并关闭 SQL 日志，避免输出敏感字段。
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	// 当前操作失败时进入错误处理分支。
	if err != nil {
		// 把当前错误返回给调用方；在事务回调中这会触发回滚。
		return err
		// 结束当前作用域的代码块或结构体定义。
	}
	// 从 GORM 取出底层 SQL 连接池。
	sqlDB, err := db.DB()
	// 当前操作失败时进入错误处理分支。
	if err != nil {
		// 把当前错误返回给调用方；在事务回调中这会触发回滚。
		return err
		// 结束当前作用域的代码块或结构体定义。
	}
	// 当前函数退出时关闭该连接或测试服务，释放资源。
	defer sqlDB.Close()
	// 最多允许同时打开 30 个数据库连接。
	sqlDB.SetMaxOpenConns(30)
	// 连接池最多保留 10 个空闲连接。
	sqlDB.SetMaxIdleConns(10)
	// 单个数据库连接最多复用 30 分钟。
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	// 创建配置好的 Redis 客户端。
	redisClient := redis.NewClient(&redis.Options{
		// 传入 Redis 地址、逻辑库编号和认证密码。
		Addr: cfg.RedisAddress, DB: cfg.RedisDB, Password: cfg.RedisPassword,
		// 设置 Redis 建连一秒、读写各半秒的超时。
		DialTimeout: time.Second, ReadTimeout: 500 * time.Millisecond, WriteTimeout: 500 * time.Millisecond,
		// 让 Redis 操作遵守上下文超时，并关闭失败后的自动重试。
		ContextTimeoutEnabled: true, MaxRetries: -1,
		// 结束事务、清理函数或子测试的回调调用。
	})
	// 当前函数退出时关闭该连接或测试服务，释放资源。
	defer redisClient.Close()
	// 启动检查使用最多两秒的独立上下文。
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	// 启动时检查 Redis；连接失败时记录降级信息。
	if redisClient.Ping(ctx).Err() != nil {
		// 输出启动或运行状态日志。
		log.Print("Redis unavailable; database queries remain available")
		// 结束当前作用域的代码块或结构体定义。
	}
	// 启动检查结束后立即释放上下文资源。
	cancel()
	// 设置 Gin 的发布模式，减少开发调试输出。
	gin.SetMode(gin.ReleaseMode)
	// 组合数据库、缓存与配置，创建业务服务实例。
	app := server.New(db, &cache.Store{Client: redisClient, Prefix: cfg.CachePrefix}, cfg)
	// 禁止 AI 请求自动跟随重定向，避免密钥被转发到意外地址。
	app.Service.HTTP.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	// 创建 HTTP 服务并限制请求头读取时间。
	listener := &http.Server{Addr: cfg.Address, Handler: app.Router(), ReadHeaderTimeout: 10 * time.Second,
		// 设置请求读取、空闲连接超时及请求头大小上限。
		ReadTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	// 监听中断和终止信号，用上下文通知服务退出。
	stopped, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// 函数退出时停止监听系统信号。
	defer stop()
	// 用带一个缓冲位的通道接收 HTTP 服务退出结果。
	errorsCh := make(chan error, 1)
	// 在独立 goroutine 中运行 HTTP 服务。
	go func() {
		// 输出启动或运行状态日志。
		log.Printf("Go backend listening at http://%s", cfg.Address)
		// 开始监听端口，并把退出或启动错误发送到通道。
		errorsCh <- listener.ListenAndServe()
		// 结束 goroutine 函数并立即启动执行。
	}()
	// 等待 HTTP 服务退出或系统终止信号。
	select {
	// 收到 HTTP 服务退出结果时进入此分支。
	case err := <-errorsCh:
		// HTTP 服务因异常退出而非正常关闭时返回错误。
		if !errors.Is(err, http.ErrServerClosed) {
			// 把当前错误返回给调用方；在事务回调中这会触发回滚。
			return err
			// 结束当前作用域的代码块或结构体定义。
		}
	// 收到系统退出信号时进入优雅关闭分支。
	case <-stopped.Done():
		// 退出时最多等待十秒，让正在执行的请求结束。
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		// 函数退出时取消上下文，释放计时器资源。
		defer cancel()
		// 停止接收新请求，并等待已接收的请求完成。
		return listener.Shutdown(ctx)
		// 结束当前作用域的代码块或结构体定义。
	}
	// 返回空错误，表示当前操作成功结束。
	return nil
	// 结束当前作用域的代码块或结构体定义。
}
