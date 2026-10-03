// 声明环境配置包。
package config

// 开始声明本文件使用的标准库、第三方库和项目内部包。
import (
	// 引入 testing，用于编写和运行 Go 测试。
	"testing"
	// 引入 time，用于处理时间、有效期和超时。
	"time"

	// 引入 github.com/go-sql-driver/mysql，用于使用 MySQL 驱动、连接配置或错误码。
	driver "github.com/go-sql-driver/mysql"
	// 结束依赖导入列表。
)

// 测试特殊密码、时间解析与时区是否正确写入连接配置。
func TestDSNPreservesSpecialPasswordAndTimezone(t *testing.T) {
	// 为连接配置测试加载上海时区。
	loc, _ := time.LoadLocation("Asia/Shanghai")
	// 构造包含特殊字符密码的测试连接配置。
	c := Config{DBHost: "127.0.0.1", DBPort: "3306", DBUser: "test", DBPassword: "test@:/#password", DBName: "test", Charset: "utf8mb4", Location: loc}
	// 重新解析连接字符串，确认参数未在格式化中丢失。
	parsed, err := driver.ParseDSN(c.DSN())
	// 当前操作失败时进入错误处理分支。
	if err != nil {
		// 记录失败原因并立即结束当前测试。
		t.Fatal(err)
		// 结束当前作用域的代码块或结构体定义。
	}
	// 确认特殊密码未变、时间解析已开启，且时区仍为上海。
	if parsed.Passwd != c.DBPassword || !parsed.ParseTime || parsed.Loc.String() != "Asia/Shanghai" {
		// 记录失败原因并立即结束当前测试。
		t.Fatal("DSN lost password or time settings")
		// 结束当前作用域的代码块或结构体定义。
	}
	// 结束当前作用域的代码块或结构体定义。
}
