package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App       AppConfig       `yaml:"app"`
	Server    ServerConfig    `yaml:"server"`
	MySQL     MySQLConfig     `yaml:"mysql"`
	Redis     RedisConfig     `yaml:"redis"`
	JWT       JWTConfig       `yaml:"jwt"`
	Log       LogConfig       `yaml:"log"`
	Worker    WorkerConfig    `yaml:"worker"`
	Cron      CronConfig      `yaml:"cron"`
	Metrics   MetricsConfig   `yaml:"metrics"`
	Migration MigrationConfig `yaml:"migration"`
	Security  SecurityConfig  `yaml:"security"`
	Tracing   TracingConfig   `yaml:"tracing"`
	Resilience ResilienceConfig `yaml:"resilience"`
}

// ResilienceConfig 限流与熔断配置
type ResilienceConfig struct {
	RateLimit      RateLimitConfig      `yaml:"rate-limit"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit-breaker"`
}

type RateLimitConfig struct {
	Enabled    bool `yaml:"enabled"`
	RPS        int  `yaml:"rps"`         // 每秒请求数
	Burst      int  `yaml:"burst"`       // 突发容量
	PerIP      bool `yaml:"per-ip"`      // 是否按 IP 限流
}

type CircuitBreakerConfig struct {
	Enabled          bool `yaml:"enabled"`
	MaxRequests      int  `yaml:"max-requests"`       // 半开状态允许的探测请求数
	Interval         int  `yaml:"interval"`           // 统计窗口秒数
	Timeout          int  `yaml:"timeout"`            // 熔断后多久进入半开（秒）
	FailureThreshold int  `yaml:"failure-threshold"`  // 失败阈值触发熔断
}

type TracingConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	Service  string `yaml:"service"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env"`
}

type ServerConfig struct {
	Host                   string `yaml:"host"`
	Port                   int    `yaml:"port"`
	ReadTimeoutSeconds     int    `yaml:"read-timeout-seconds"`
	WriteTimeoutSeconds    int    `yaml:"write-timeout-seconds"`
	ShutdownTimeoutSeconds int    `yaml:"shutdown-timeout-seconds"`
}

type MySQLConfig struct {
	Path         string `yaml:"path"`
	DBName       string `yaml:"db-name"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Config       string `yaml:"config"`
	MaxIdleConns int    `yaml:"max-idle-conns"`
	MaxOpenConns int    `yaml:"max-open-conns"`
	LogMode      bool   `yaml:"log-mode"`
}

type RedisConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Addr         string `yaml:"addr"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool-size"`
	MinIdleConns int    `yaml:"min-idle-conns"`
	MaxRetries   int    `yaml:"max-retries"`
	DialTimeout  int    `yaml:"dial-timeout"`
	ReadTimeout  int    `yaml:"read-timeout"`
	WriteTimeout int    `yaml:"write-timeout"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire-hours"`
	Issuer      string `yaml:"issuer"`
}

type LogConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max-size"`
	MaxBackups int    `yaml:"max-backups"`
	MaxAge     int    `yaml:"max-age"`
	Compress   bool   `yaml:"compress"`
}

type WorkerConfig struct {
	Enabled                bool   `yaml:"enabled"`
	PoolSize               int    `yaml:"pool-size"`
	ConsumerGroup          string `yaml:"consumer-group"`
	BatchSize              int    `yaml:"batch-size"`
	BatchTimeoutSeconds    int    `yaml:"batch-timeout-seconds"`
	BlockTimeoutSeconds    int    `yaml:"block-timeout-seconds"`
	ReclaimIntervalSeconds int    `yaml:"reclaim-interval-seconds"`
	MinIdleSeconds         int    `yaml:"min-idle-seconds"`
	MaxRetry               int    `yaml:"max-retry"`
	DeadLetterStream       string `yaml:"dead-letter-stream"`
	ConsumerPrefix         string `yaml:"consumer-prefix"`
}

type CronConfig struct {
	Enabled            bool `yaml:"enabled"`
	UseDistributedLock bool `yaml:"use-distributed-lock"`
	LockTTLSeconds     int  `yaml:"lock-ttl-seconds"`
}

type MetricsConfig struct {
	Enabled bool `yaml:"enabled"`
}

type MigrationConfig struct {
	AutoMigrate bool `yaml:"auto-migrate"`
}

type SecurityConfig struct {
	CORSAllowedOrigins []string `yaml:"cors-allowed-origins"`
	SeedDefaultUsers   bool     `yaml:"seed-default-users"`
}

func Load(path string) (*Config, error) {
	cfg := defaultConfig()
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(content, cfg); err != nil {
		return nil, err
	}
	cfg.applyEnvOverrides()
	cfg.applyDerivedDefaults()
	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name: "bookadmin",
			Env:  "development",
		},
		Server: ServerConfig{
			Host:                   "0.0.0.0",
			Port:                   8888,
			ReadTimeoutSeconds:     10,
			WriteTimeoutSeconds:    10,
			ShutdownTimeoutSeconds: 10,
		},
		MySQL: MySQLConfig{
			Path:         "127.0.0.1:3306",
			DBName:       "bookadmin",
			Username:     "root",
			Password:     "root",
			Config:       "charset=utf8mb4&parseTime=True&loc=Local",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			LogMode:      true,
		},
		Redis: RedisConfig{
			Enabled:      true,
			Addr:         "127.0.0.1:6379",
			Password:     "",
			DB:           0,
			PoolSize:     100,
			MinIdleConns: 10,
			MaxRetries:   3,
			DialTimeout:  5,
			ReadTimeout:  3,
			WriteTimeout: 3,
		},
		JWT: JWTConfig{
			Secret:      "bookadmin-secret-key-change-in-production",
			ExpireHours: 24,
			Issuer:      "bookadmin",
		},
		Log: LogConfig{
			Level:      "info",
			File:       "./log/app.log",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   false,
		},
		Worker: WorkerConfig{
			Enabled:                true,
			PoolSize:               5,
			ConsumerGroup:          "sync-group",
			BatchSize:              100,
			BatchTimeoutSeconds:    5,
			BlockTimeoutSeconds:    1,
			ReclaimIntervalSeconds: 15,
			MinIdleSeconds:         30,
			MaxRetry:               5,
			DeadLetterStream:       "stream:dead-letter:sync",
			ConsumerPrefix:         "worker",
		},
		Cron: CronConfig{
			Enabled:            true,
			UseDistributedLock: true,
			LockTTLSeconds:     600,
		},
		Metrics: MetricsConfig{
			Enabled: true,
		},
		Migration: MigrationConfig{
			AutoMigrate: true,
		},
		Security: SecurityConfig{
			CORSAllowedOrigins: []string{"http://localhost:8080"},
			SeedDefaultUsers:   true,
		},
		Tracing: TracingConfig{
			Enabled:  true,
			Endpoint: "localhost:4317",
			Service:  "bookadmin-api",
		},
		Resilience: ResilienceConfig{
			RateLimit: RateLimitConfig{
				Enabled: true,
				RPS:     100,
				Burst:   50,
				PerIP:   true,
			},
			CircuitBreaker: CircuitBreakerConfig{
				Enabled:          true,
				MaxRequests:      3,
				Interval:         10,
				Timeout:          30,
				FailureThreshold: 5,
			},
		},
	}
}

func (c *Config) applyEnvOverrides() {
	applyString("APP_NAME", &c.App.Name)
	applyString("APP_ENV", &c.App.Env)
	applyString("SERVER_HOST", &c.Server.Host)
	applyInt("SERVER_PORT", &c.Server.Port)
	applyInt("SERVER_READ_TIMEOUT_SECONDS", &c.Server.ReadTimeoutSeconds)
	applyInt("SERVER_WRITE_TIMEOUT_SECONDS", &c.Server.WriteTimeoutSeconds)
	applyInt("SERVER_SHUTDOWN_TIMEOUT_SECONDS", &c.Server.ShutdownTimeoutSeconds)

	applyString("MYSQL_PATH", &c.MySQL.Path)
	applyString("MYSQL_DB_NAME", &c.MySQL.DBName)
	applyString("MYSQL_USERNAME", &c.MySQL.Username)
	applyString("MYSQL_PASSWORD", &c.MySQL.Password)
	applyString("MYSQL_CONFIG", &c.MySQL.Config)
	applyInt("MYSQL_MAX_IDLE_CONNS", &c.MySQL.MaxIdleConns)
	applyInt("MYSQL_MAX_OPEN_CONNS", &c.MySQL.MaxOpenConns)
	applyBool("MYSQL_LOG_MODE", &c.MySQL.LogMode)

	applyBool("REDIS_ENABLED", &c.Redis.Enabled)
	applyString("REDIS_ADDR", &c.Redis.Addr)
	applyString("REDIS_PASSWORD", &c.Redis.Password)
	applyInt("REDIS_DB", &c.Redis.DB)
	applyInt("REDIS_POOL_SIZE", &c.Redis.PoolSize)
	applyInt("REDIS_MIN_IDLE_CONNS", &c.Redis.MinIdleConns)
	applyInt("REDIS_MAX_RETRIES", &c.Redis.MaxRetries)
	applyInt("REDIS_DIAL_TIMEOUT_SECONDS", &c.Redis.DialTimeout)
	applyInt("REDIS_READ_TIMEOUT_SECONDS", &c.Redis.ReadTimeout)
	applyInt("REDIS_WRITE_TIMEOUT_SECONDS", &c.Redis.WriteTimeout)

	applyString("JWT_SECRET", &c.JWT.Secret)
	applyInt("JWT_EXPIRE_HOURS", &c.JWT.ExpireHours)
	applyString("JWT_ISSUER", &c.JWT.Issuer)

	applyString("LOG_LEVEL", &c.Log.Level)
	applyString("LOG_FILE", &c.Log.File)
	applyInt("LOG_MAX_SIZE", &c.Log.MaxSize)
	applyInt("LOG_MAX_BACKUPS", &c.Log.MaxBackups)
	applyInt("LOG_MAX_AGE", &c.Log.MaxAge)
	applyBool("LOG_COMPRESS", &c.Log.Compress)

	applyBool("WORKER_ENABLED", &c.Worker.Enabled)
	applyInt("WORKER_POOL_SIZE", &c.Worker.PoolSize)
	applyString("WORKER_CONSUMER_GROUP", &c.Worker.ConsumerGroup)
	applyInt("WORKER_BATCH_SIZE", &c.Worker.BatchSize)
	applyInt("WORKER_BATCH_TIMEOUT_SECONDS", &c.Worker.BatchTimeoutSeconds)
	applyInt("WORKER_BLOCK_TIMEOUT_SECONDS", &c.Worker.BlockTimeoutSeconds)
	applyInt("WORKER_RECLAIM_INTERVAL_SECONDS", &c.Worker.ReclaimIntervalSeconds)
	applyInt("WORKER_MIN_IDLE_SECONDS", &c.Worker.MinIdleSeconds)
	applyInt("WORKER_MAX_RETRY", &c.Worker.MaxRetry)
	applyString("WORKER_DEAD_LETTER_STREAM", &c.Worker.DeadLetterStream)
	applyString("WORKER_CONSUMER_PREFIX", &c.Worker.ConsumerPrefix)

	applyBool("CRON_ENABLED", &c.Cron.Enabled)
	applyBool("CRON_USE_DISTRIBUTED_LOCK", &c.Cron.UseDistributedLock)
	applyInt("CRON_LOCK_TTL_SECONDS", &c.Cron.LockTTLSeconds)

	applyBool("METRICS_ENABLED", &c.Metrics.Enabled)
	applyBool("MIGRATION_AUTO_MIGRATE", &c.Migration.AutoMigrate)
	applyBool("SECURITY_SEED_DEFAULT_USERS", &c.Security.SeedDefaultUsers)

	applyBool("TRACING_ENABLED", &c.Tracing.Enabled)
	applyString("OTEL_EXPORTER_OTLP_ENDPOINT", &c.Tracing.Endpoint)
	applyString("OTEL_SERVICE_NAME", &c.Tracing.Service)

	applyBool("RATE_LIMIT_ENABLED", &c.Resilience.RateLimit.Enabled)
	applyInt("RATE_LIMIT_RPS", &c.Resilience.RateLimit.RPS)
	applyInt("RATE_LIMIT_BURST", &c.Resilience.RateLimit.Burst)
	applyBool("RATE_LIMIT_PER_IP", &c.Resilience.RateLimit.PerIP)
	applyBool("CIRCUIT_BREAKER_ENABLED", &c.Resilience.CircuitBreaker.Enabled)
	applyInt("CIRCUIT_BREAKER_MAX_REQUESTS", &c.Resilience.CircuitBreaker.MaxRequests)
	applyInt("CIRCUIT_BREAKER_INTERVAL", &c.Resilience.CircuitBreaker.Interval)
	applyInt("CIRCUIT_BREAKER_TIMEOUT", &c.Resilience.CircuitBreaker.Timeout)
	applyInt("CIRCUIT_BREAKER_FAILURE_THRESHOLD", &c.Resilience.CircuitBreaker.FailureThreshold)

	if raw := strings.TrimSpace(os.Getenv("SECURITY_CORS_ALLOWED_ORIGINS")); raw != "" {
		parts := strings.Split(raw, ",")
		c.Security.CORSAllowedOrigins = c.Security.CORSAllowedOrigins[:0]
		for _, part := range parts {
			if origin := strings.TrimSpace(part); origin != "" {
				c.Security.CORSAllowedOrigins = append(c.Security.CORSAllowedOrigins, origin)
			}
		}
	}
}

func (c *Config) applyDerivedDefaults() {
	if c.App.Name == "" {
		c.App.Name = "bookadmin"
	}
	if c.App.Env == "" {
		c.App.Env = "development"
	}
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8888
	}
	if c.Server.ReadTimeoutSeconds <= 0 {
		c.Server.ReadTimeoutSeconds = 10
	}
	if c.Server.WriteTimeoutSeconds <= 0 {
		c.Server.WriteTimeoutSeconds = 10
	}
	if c.Server.ShutdownTimeoutSeconds <= 0 {
		c.Server.ShutdownTimeoutSeconds = 10
	}
	if c.JWT.ExpireHours <= 0 {
		c.JWT.ExpireHours = 24
	}
	if c.JWT.Issuer == "" {
		c.JWT.Issuer = c.App.Name
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.File == "" {
		c.Log.File = "./log/app.log"
	}
	if c.Worker.ConsumerGroup == "" {
		c.Worker.ConsumerGroup = "sync-group"
	}
	if c.Worker.PoolSize <= 0 {
		c.Worker.PoolSize = 1
	}
	if c.Worker.BatchSize <= 0 {
		c.Worker.BatchSize = 100
	}
	if c.Worker.BatchTimeoutSeconds <= 0 {
		c.Worker.BatchTimeoutSeconds = 5
	}
	if c.Worker.BlockTimeoutSeconds <= 0 {
		c.Worker.BlockTimeoutSeconds = 1
	}
	if c.Worker.ReclaimIntervalSeconds <= 0 {
		c.Worker.ReclaimIntervalSeconds = 15
	}
	if c.Worker.MinIdleSeconds <= 0 {
		c.Worker.MinIdleSeconds = 30
	}
	if c.Worker.MaxRetry <= 0 {
		c.Worker.MaxRetry = 5
	}
	if c.Worker.DeadLetterStream == "" {
		c.Worker.DeadLetterStream = "stream:dead-letter:sync"
	}
	if c.Worker.ConsumerPrefix == "" {
		c.Worker.ConsumerPrefix = "worker"
	}
	if len(c.Security.CORSAllowedOrigins) == 0 {
		c.Security.CORSAllowedOrigins = []string{"http://localhost:8080"}
	}
	if c.Tracing.Service == "" {
		c.Tracing.Service = c.App.Name
	}
	if c.Tracing.Endpoint == "" {
		c.Tracing.Endpoint = "localhost:4317"
	}
	if c.Resilience.RateLimit.RPS <= 0 {
		c.Resilience.RateLimit.RPS = 100
	}
	if c.Resilience.RateLimit.Burst <= 0 {
		c.Resilience.RateLimit.Burst = 50
	}
	if c.Resilience.CircuitBreaker.Interval <= 0 {
		c.Resilience.CircuitBreaker.Interval = 10
	}
	if c.Resilience.CircuitBreaker.Timeout <= 0 {
		c.Resilience.CircuitBreaker.Timeout = 30
	}
	if c.Resilience.CircuitBreaker.FailureThreshold == 0 {
		c.Resilience.CircuitBreaker.FailureThreshold = 5
	}
	if c.Resilience.CircuitBreaker.MaxRequests == 0 {
		c.Resilience.CircuitBreaker.MaxRequests = 3
	}
}

func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.App.Env, "production")
}

func (c *Config) ServerAddress() string {
	return c.Server.Host + ":" + strconv.Itoa(c.Server.Port)
}

func (c *Config) JWTDuration() time.Duration {
	return time.Duration(c.JWT.ExpireHours) * time.Hour
}

func applyString(key string, target *string) {
	if value, ok := os.LookupEnv(key); ok {
		*target = value
	}
}

func applyInt(key string, target *int) {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(value); err == nil {
			*target = parsed
		}
	}
}

func applyBool(key string, target *bool) {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.ParseBool(value); err == nil {
			*target = parsed
		}
	}
}
