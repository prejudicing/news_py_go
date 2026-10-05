package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/prejudicing/news_py_go/backend_go/internal/cache"
	"github.com/prejudicing/news_py_go/backend_go/internal/config"
	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	"github.com/prejudicing/news_py_go/backend_go/internal/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 封装集成测试的请求发送与响应断言。
type testAPI struct {
	t             *testing.T
	router        http.Handler
	token         string
	refreshCookie string
	lastCookie    *http.Cookie
}

// 发送测试请求并检查 HTTP 状态与响应格式。
func (a *testAPI) request(method, path string, data any, status int) map[string]any {
	a.t.Helper()
	body, _ := json.Marshal(data)
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if a.token != "" {
		req.Header.Set("Authorization", a.token)
	}
	if a.refreshCookie != "" {
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: a.refreshCookie})
	}
	w := httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	if w.Code != status {
		a.t.Fatalf("%s %s: got %d want %d: %s", method, path, w.Code, status, w.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		a.t.Fatal(err)
	}
	if result["code"] != float64(status) {
		a.t.Fatal("response envelope mismatch")
	}
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "refresh_token" {
			a.lastCookie = cookie
			if cookie.MaxAge < 0 {
				a.refreshCookie = ""
			} else {
				a.refreshCookie = cookie.Value
			}
		}
	}
	if value, valid := result["data"].(map[string]any); valid {
		return value
	}
	return result
}

// 使用隔离的 MySQL 测试库验证完整业务和刷新会话轮换。
func TestMySQLIntegration(t *testing.T) {
	if os.Getenv("GO_INTEGRATION_TEST") != "1" {
		t.Skip("set GO_INTEGRATION_TEST=1 and ENV_FILE to enable isolated MySQL tests")
	}
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "integration-test-signing-secret-with-32-bytes")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.DBName = ""
	admin, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	database := fmt.Sprintf("news_go_test_%d", time.Now().UnixNano())
	if !regexp.MustCompile(`^news_go_test_[0-9]+$`).MatchString(database) {
		t.Fatal("unsafe test database name")
	}
	if _, err := admin.Exec("CREATE DATABASE `" + database + "` CHARACTER SET utf8mb4"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, err := sql.Open("mysql", cfg.DSN())
		if err != nil {
			t.Error(err)
			return
		}
		defer cleanup.Close()
		if _, err := cleanup.Exec("DROP DATABASE `" + database + "`"); err != nil {
			t.Error(err)
		}
	})
	cfg.DBName = database
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	ddl, err := os.ReadFile("../../../docs/02-数据库sql文件/database.sql")
	if err != nil {
		t.Fatal(err)
	}
	statements := regexp.MustCompile("(?s)CREATE TABLE IF NOT EXISTS `.*?;").FindAllString(string(ddl), -1)
	if len(statements) != 8 {
		t.Fatal("unexpected source database schema")
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Exec("CREATE TABLE user_token (id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, token VARCHAR(255) NOT NULL)").Error; err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile("../../../docs/02-数据库sql文件/migrations/001_refresh_sessions.sql")
	if err != nil {
		t.Fatal(err)
	}
	migrationStatements := regexp.MustCompile("(?s)CREATE TABLE IF NOT EXISTS `refresh_sessions`.*?;|DROP TABLE IF EXISTS `user_token`;").FindAllString(string(migration), -1)
	if len(migrationStatements) != 2 {
		t.Fatal("unexpected refresh-session migration")
	}
	for _, statement := range migrationStatements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	if !db.Migrator().HasTable(&model.RefreshSession{}) || db.Migrator().HasTable("user_token") {
		t.Fatal("refresh-session migration did not replace the legacy token table")
	}
	category := model.Category{ID: 1, Name: "头条", SortOrder: 1}
	if err := db.Create(&category).Error; err != nil {
		t.Fatal(err)
	}
	for _, id := range []uint64{11, 12} {
		n := model.News{ID: id, Title: fmt.Sprintf("测试新闻%d", id), CategoryID: 1, Content: "测试正文", PublishTime: time.Now().In(cfg.Location)}
		if err := db.Create(&n).Error; err != nil {
			t.Fatal(err)
		}
	}
	r := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: r.Addr(), MaxRetries: -1, ContextTimeoutEnabled: true})
	defer redisClient.Close()
	cfg.AIKey = "server-only-test-key"
	app := New(db, &cache.Store{Client: redisClient, Prefix: "integration:"}, cfg)
	api := testAPI{t: t, router: app.Router()}
	creds := map[string]string{"username": "go-test-user", "password": "test-password"}

	t.Run("registration_login_and_refresh_reuse_detection", func(t *testing.T) {
		api.t = t
		data := api.request("POST", "/api/user/register", creds, 200)
		api.token = data["token"].(string)
		if api.refreshCookie == "" || api.lastCookie == nil || !api.lastCookie.HttpOnly || api.lastCookie.SameSite != http.SameSiteStrictMode {
			t.Fatal("refresh credential cookie flags are missing")
		}
		if _, err := utils.ParseJWT(api.token, []byte(cfg.JWTSecret)); err != nil {
			t.Fatal("registration did not issue a valid JWT")
		}
		if strings.Contains(fmt.Sprint(data), "password") {
			t.Fatal("password exposed")
		}
		api.request("POST", "/api/user/register", creds, 400)
		api.request("POST", "/api/user/login", map[string]string{"username": "go-test-user", "password": "wrong"}, 401)
		login := api.request("POST", "/api/user/login", creds, 200)
		newToken := login["token"].(string)
		api.token = ""
		api.request("GET", "/api/user/info", nil, 401)
		tampered := testAPI{t: t, router: api.router, token: newToken + "x"}
		tampered.request("GET", "/api/user/info", nil, 401)
		api.token = "Bearer " + newToken
		info := api.request("GET", "/api/user/info", nil, 200)
		if info["username"] != creds["username"] {
			t.Fatal("incorrect user")
		}
		api.token = newToken
		oldRefresh := api.refreshCookie
		refreshed := api.request("POST", "/api/user/refresh", nil, 200)
		if refreshed["token"] == newToken || api.refreshCookie == oldRefresh {
			t.Fatal("refresh did not rotate the access and refresh credentials")
		}
		rotatedRefresh := api.refreshCookie
		api.refreshCookie = oldRefresh
		api.request("POST", "/api/user/refresh", nil, 401)
		api.refreshCookie = rotatedRefresh
		api.request("POST", "/api/user/refresh", nil, 401)
	})
	t.Run("password_hash_login", func(t *testing.T) {
		u := model.User{Username: "existing-user", Password: "$2b$12$.Nujxqtcickr7smb1OWO/.aZlZTSZ1z2O4KnRJ8AXmJXDlOzM6ZIG"}
		if err := db.Create(&u).Error; err != nil {
			t.Fatal(err)
		}
		passwordAPI := &testAPI{t: t, router: api.router}
		login := passwordAPI.request("POST", "/api/user/login", map[string]string{"username": u.Username, "password": "legacy-password"}, 200)
		passwordAPI.token = login["token"].(string)
		passwordAPI.request("GET", "/api/user/info", nil, 200)
	})
	t.Run("profile_and_password", func(t *testing.T) {
		api.t = t
		profile := api.request("PUT", "/api/user/update", map[string]string{"nickname": "昵称", "bio": "简介", "gender": "female"}, 200)
		if profile["nickname"] != "昵称" {
			t.Fatal("profile not updated")
		}
		api.request("PUT", "/api/user/update", map[string]string{}, 200)
		api.request("PUT", "/api/user/update", map[string]string{"gender": "invalid"}, 422)
		api.request("PUT", "/api/user/password", map[string]string{"oldPassword": "wrong", "newPassword": "new-password"}, 400)
		api.request("PUT", "/api/user/password", map[string]string{"oldPassword": "test-password", "newPassword": "new-password"}, 200)
		api.request("POST", "/api/user/login", creds, 401)
		data := api.request("POST", "/api/user/login", map[string]string{"username": "go-test-user", "password": "new-password"}, 200)
		api.token = data["token"].(string)
	})
	t.Run("news_pagination_cache_and_views", func(t *testing.T) {
		api.t = t
		api.request("GET", "/api/news/categories", nil, 200)
		first := api.request("GET", "/api/news/list?categoryId=1&pageSize=1", nil, 200)
		if first["hasMore"] != true || first["total"] != float64(2) {
			t.Fatal("incorrect pagination")
		}
		item := first["list"].([]any)[0].(map[string]any)
		if item["publishTime"] == nil || item["categoryId"] != float64(1) {
			t.Fatal("frontend fields missing")
		}
		api.request("GET", "/api/news/list?categoryId=1&pageSize=1", nil, 200)
		if !r.Exists("integration:list:1:1:1") {
			t.Fatal("Redis cache not populated")
		}
		second := api.request("GET", "/api/news/list?categoryId=1&pageSize=1&page=2", nil, 200)
		if second["hasMore"] != false {
			t.Fatal("incorrect last page")
		}
		detail := api.request("GET", "/api/news/detail?id=11", nil, 200)
		related := detail["relatedNews"].([]any)
		if len(related) != 1 || related[0].(map[string]any)["id"] != float64(12) {
			t.Fatal("related news must match the category and exclude the current item")
		}
		cached := api.request("GET", "/api/news/detail?id=11", nil, 200)
		if cached["views"].(float64) != detail["views"].(float64)+1 {
			t.Fatal("cached views did not increase")
		}
		api.request("GET", "/api/news/detail?id=999", nil, 404)
		r.Close()
		api.request("GET", "/api/news/list?categoryId=1", nil, 200)
	})
	t.Run("favorites", func(t *testing.T) {
		api.t = t
		api.request("POST", "/api/favorite/add", map[string]int{"newsId": 11}, 200)
		api.request("POST", "/api/favorite/add", map[string]int{"newsId": 11}, 400)
		if api.request("GET", "/api/favorite/check?newsId=11", nil, 200)["isFavorite"] != true {
			t.Fatal("favorite missing")
		}
		list := api.request("GET", "/api/favorite/list", nil, 200)
		item := list["list"].([]any)[0].(map[string]any)
		if item["favoriteId"] == nil || item["favoriteTime"] == nil || item["publishTime"] == nil {
			t.Fatal("favorite fields missing")
		}
		api.request("DELETE", "/api/favorite/remove?newsId=11", nil, 200)
		api.request("DELETE", "/api/favorite/remove?newsId=11", nil, 404)
		api.request("POST", "/api/favorite/add", map[string]int{"newsId": 12}, 200)
		api.request("DELETE", "/api/favorite/clear", nil, 200)
		if api.request("GET", "/api/favorite/list", nil, 200)["total"] != float64(0) {
			t.Fatal("favorites not cleared")
		}
	})
	t.Run("history_record_ids_and_user_isolation", func(t *testing.T) {
		api.t = t
		first := api.request("POST", "/api/history/add", map[string]int{"newsId": 11}, 200)
		repeat := api.request("POST", "/api/history/add", map[string]int{"newsId": 11}, 200)
		if first["id"] != repeat["id"] {
			t.Fatal("duplicate history created")
		}
		api.request("POST", "/api/history/add", map[string]int{"newsId": 12}, 200)
		list := api.request("GET", "/api/history/list", nil, 200)
		if list["total"] != float64(2) {
			t.Fatal("incorrect history count")
		}
		item := list["list"].([]any)[0].(map[string]any)
		if item["historyId"] == nil || item["viewTime"] == nil || item["publishTime"] == nil {
			t.Fatal("history fields missing")
		}
		other := testAPI{t: t, router: api.router}
		other.token = other.request("POST", "/api/user/register", map[string]string{"username": "other-user", "password": "test-password"}, 200)["token"].(string)
		path := fmt.Sprintf("/api/history/delete/%.0f", first["id"].(float64))
		other.request("DELETE", path, nil, 404)
		api.request("DELETE", path, nil, 200)
		if api.request("GET", "/api/history/list", nil, 200)["total"] != float64(1) {
			t.Fatal("wrong history deleted")
		}
		api.request("DELETE", "/api/history/clear", nil, 200)
	})
	t.Run("ai_authenticated_stream_and_sanitized_errors", func(t *testing.T) {
		api.t = t
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Authorization") != "Bearer server-only-test-key" {
				t.Error("wrong server credentials")
			}
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\ndata: [DONE]\n\n"))
		}))
		defer upstream.Close()
		app.Services.AI.Config.AIEndpoint = upstream.URL
		app.Services.AI.HTTP = upstream.Client()
		body := `{"messages":[{"role":"user","content":"hello"}]}`
		req := httptest.NewRequest("POST", "/api/ai/chat", strings.NewReader(body))
		req.Header.Set("Authorization", api.token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		api.router.ServeHTTP(w, req)
		if w.Code != 200 || !strings.Contains(w.Body.String(), "[DONE]") || strings.Contains(w.Body.String(), app.Services.AI.Config.AIKey) {
			t.Fatal("invalid AI stream")
		}
		app.Services.AI.Config.AIKey = ""
		api.request("POST", "/api/ai/chat", map[string]any{"messages": []dto.ChatMessage{{Role: "user", Content: "hello"}}}, 503)
		app.Services.AI.Config.AIKey = "server-only-test-key"
		failed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(401)
			_, _ = w.Write([]byte(app.Services.AI.Config.AIKey))
		}))
		defer failed.Close()
		app.Services.AI.Config.AIEndpoint = failed.URL
		api.request("POST", "/api/ai/chat", map[string]any{"messages": []dto.ChatMessage{{Role: "user", Content: "hello"}}}, 502)
	})
}
