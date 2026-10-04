package config

import (
	"testing"
	"time"

	driver "github.com/go-sql-driver/mysql"
)

// 测试特殊密码、时间解析与时区是否正确写入连接配置。
func TestDSNPreservesSpecialPasswordAndTimezone(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	c := Config{DBHost: "127.0.0.1", DBPort: "3306", DBUser: "test", DBPassword: "test@:/#password", DBName: "test", Charset: "utf8mb4", Location: loc}
	parsed, err := driver.ParseDSN(c.DSN())
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Passwd != c.DBPassword || !parsed.ParseTime || parsed.Loc.String() != "Asia/Shanghai" {
		t.Fatal("DSN lost password or time settings")
	}
}
