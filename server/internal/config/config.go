package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构体
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Log      LogConfig
	Env      string
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  int
	WriteTimeout int
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string
	Expiration int
}

// LogConfig 日志配置
type LogConfig struct {
	Level string
	File  string
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	// 确定环境
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// 初始化 viper
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 配置文件路径
	configPaths := []string{
		".",
		"./config",
		"./internal/config",
		"./../config",
		"./../internal/config",
	}

	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	// 配置文件名（不带扩展名）
	v.SetConfigName(fmt.Sprintf("config.%s", env))

	// 配置文件类型
	v.SetConfigType("yaml")

	// 读取环境变量
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 尝试读取配置文件
	err := v.ReadInConfig()
	if err != nil {
		// 配置文件不存在时不报错，使用默认值和环境变量
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 构建配置
	config := &Config{
		Env: env,
		Server: ServerConfig{
			Port:         getEnvOrConfig(v, "SERVER_PORT", "server.port"),
			Host:         getEnvOrConfig(v, "SERVER_HOST", "server.host"),
			ReadTimeout:  v.GetInt("server.readTimeout"),
			WriteTimeout: v.GetInt("server.writeTimeout"),
		},
		Database: DatabaseConfig{
			Host:     getEnvOrConfig(v, "DB_HOST", "database.host"),
			Port:     getEnvOrConfig(v, "DB_PORT", "database.port"),
			User:     getEnvOrConfig(v, "DB_USER", "database.user"),
			Password: getEnvOrConfig(v, "DB_PASSWORD", "database.password"),
			DBName:   getEnvOrConfig(v, "DB_NAME", "database.dbname"),
			SSLMode:  getEnvOrConfig(v, "DB_SSLMODE", "database.sslmode"),
		},
		JWT: JWTConfig{
			Secret:     getEnvOrConfig(v, "JWT_SECRET", "jwt.secret"),
			Expiration: v.GetInt("jwt.expiration"),
		},
		Log: LogConfig{
			Level: getEnvOrConfig(v, "LOG_LEVEL", "log.level"),
			File:  getEnvOrConfig(v, "LOG_FILE", "log.file"),
		},
		Redis: RedisConfig{
			Host:     getEnvOrConfig(v, "REDIS_HOST", "redis.host"),
			Port:     getEnvOrConfig(v, "REDIS_PORT", "redis.port"),
			Password: getEnvOrConfig(v, "REDIS_PASSWORD", "redis.password"),
			DB:       v.GetInt("redis.db"),
			PoolSize: v.GetInt("redis.poolSize"),
		},
	}

	return config, nil
}

// setDefaults 设置默认配置值
func setDefaults(v *viper.Viper) {
	// 服务器默认配置
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.readTimeout", 15)
	v.SetDefault("server.writeTimeout", 15)

	// 数据库默认配置
	v.SetDefault("database.host", "127.0.0.1")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.dbname", "hermes_platform")
	v.SetDefault("database.sslmode", "disable")

	// JWT默认配置
	v.SetDefault("jwt.secret", "your-secret-key")
	v.SetDefault("jwt.expiration", 24)

	// 日志默认配置
	v.SetDefault("log.level", "info")
	v.SetDefault("log.file", "")

	// Redis默认配置
	v.SetDefault("redis.host", "127.0.0.1")
	v.SetDefault("redis.port", "6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.poolSize", 10)
}

// getEnvOrConfig 获取环境变量或配置值
func getEnvOrConfig(v *viper.Viper, envKey, configKey string) string {
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	return v.GetString(configKey)
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}
