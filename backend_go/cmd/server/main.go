package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prejudicing/news_py_go/backend_go/internal/cache"
	"github.com/prejudicing/news_py_go/backend_go/internal/config"
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	"github.com/prejudicing/news_py_go/backend_go/internal/server"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 程序入口，启动服务并处理启动失败。
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// 初始化依赖，启动 HTTP 服务并处理退出信号。
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return err
	}
	if !db.Migrator().HasTable(&model.RefreshSession{}) {
		return errors.New("required table refresh_sessions is missing; apply migration 001_refresh_sessions.sql")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress, DB: cfg.RedisDB, Password: cfg.RedisPassword,
		DialTimeout: time.Second, ReadTimeout: 500 * time.Millisecond, WriteTimeout: 500 * time.Millisecond,
		ContextTimeoutEnabled: true, MaxRetries: -1,
	})
	defer redisClient.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	if redisClient.Ping(ctx).Err() != nil {
		log.Print("Redis unavailable; database queries remain available")
	}
	cancel()
	gin.SetMode(gin.ReleaseMode)
	app := server.New(db, &cache.Store{Client: redisClient, Prefix: cfg.CachePrefix}, cfg)
	listener := &http.Server{Addr: cfg.Address, Handler: app.Router(), ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	stopped, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errorsCh := make(chan error, 1)
	go func() {
		log.Printf("Go backend listening at http://%s", cfg.Address)
		errorsCh <- listener.ListenAndServe()
	}()
	select {
	case err := <-errorsCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-stopped.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return listener.Shutdown(ctx)
	}
	return nil
}
