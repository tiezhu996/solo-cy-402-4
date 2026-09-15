package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 集中解析全部环境变量配置。
type Config struct {
	AppEnv             string
	ServerPort         string
	DBHost             string
	DBPort             string
	DBName             string
	DBUser             string
	DBPassword         string
	JWTSecret          string
	JWTExpireHours     int
	RateLimitPerMinute int
	UploadDir          string
	UploadMaxMB        int64
	CORSOrigins        []string
}

// Load 从环境变量加载配置，未设置时使用默认值。
func Load() *Config {
	return &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		DBHost:             getEnv("DB_HOST", "127.0.0.1"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBName:             getEnv("DB_NAME", "cylawcase_db"),
		DBUser:             getEnv("DB_USER", "cylawcase_user"),
		DBPassword:         getEnv("DB_PASSWORD", "cylawcase_pwd"),
		JWTSecret:          getEnv("JWT_SECRET", "change_me_to_a_long_random_string"),
		JWTExpireHours:     getEnvInt("JWT_EXPIRE_HOURS", 72),
		RateLimitPerMinute: getEnvInt("RATE_LIMIT_PER_MINUTE", 120),
		UploadDir:          getEnv("UPLOAD_DIR", "./uploads"),
		UploadMaxMB:        int64(getEnvInt("UPLOAD_MAX_MB", 10)),
		CORSOrigins:        parseCSV(getEnv("APP_CORS_ORIGINS", "http://localhost:28031")),
	}
}

// DBDSN 返回 GORM PostgreSQL DSN。
func (c *Config) DBDSN() string {
	return "host=" + c.DBHost + " user=" + c.DBUser + " password=" + c.DBPassword +
		" dbname=" + c.DBName + " port=" + c.DBPort + " sslmode=disable TimeZone=Asia/Shanghai"
}

// JWTExpireDuration 返回 JWT 过期时长。
func (c *Config) JWTExpireDuration() time.Duration {
	return time.Duration(c.JWTExpireHours) * time.Hour
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func parseCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}
