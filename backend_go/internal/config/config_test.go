package config

import (
	"os"
	"path/filepath"
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

// TestLoadRequiresJWTSecret 确认启动配置拒绝缺失或过短的签名密钥。
func TestLoadRequiresJWTSecret(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "empty.env")
	if err := os.WriteFile(configFile, nil, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ENV_FILE", configFile)
	for _, secret := range []string{"", "short", "replace-with-a-random-secret-of-at-least-32-bytes"} {
		t.Setenv("JWT_SECRET", secret)
		if _, err := Load(); err == nil {
			t.Fatal("invalid JWT signing secret accepted")
		}
	}
	t.Setenv("JWT_SECRET", "local-test-secret-with-at-least-32-bytes")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
}
