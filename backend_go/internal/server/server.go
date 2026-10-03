// 声明应用依赖装配包。
package server

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入github.com/gin-gonic/gin，用于Gin 路由和请求上下文。
	"github.com/gin-gonic/gin"
	// 引入gorm.io/gorm，用于GORM 查询与事务。
	"gorm.io/gorm"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/cache，用于Redis 缓存封装。
	"github.com/prejudicing/news_py_go/backend_go/internal/cache"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/config，用于环境配置。
	"github.com/prejudicing/news_py_go/backend_go/internal/config"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/handler，用于HTTP 请求处理。
	"github.com/prejudicing/news_py_go/backend_go/internal/handler"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/repository，用于仓储接口和统一错误。
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/service，用于业务逻辑。
	"github.com/prejudicing/news_py_go/backend_go/internal/service"
	// 结束当前声明或代码作用域。
)

// Server 是依赖装配入口，业务实现在独立分层中。
type Server struct{ Service *service.Service }

// 创建实例并注入所需依赖。
func New(db *gorm.DB, cached *cache.Store, cfg config.Config) *Server {
	// 创建实例并注入所需依赖并保存操作结果。
	repo := repository.New(db)
	// 创建实例并注入所需依赖，把结果或错误返回给调用方。
	return &Server{Service: service.New(repo, cached, cfg)}
	// 结束当前声明或代码作用域。
}

// 注册 HTTP 路由与通用中间件。
func (s *Server) Router() *gin.Engine { return handler.New(s.Service).Router() }
