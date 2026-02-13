package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// GlobalConfig holds the application configuration.
var GlobalConfig *Config

// Deprecated: Use GlobalConfig.App.JwtSecret or DI
var JwtSecret string

// Deprecated: Use GlobalConfig.Database...
var (
	DbHost                   string
	DbPort                   string
	DbUser                   string
	DbPassword               string
	DbName                   string
	DbSSLMode                string
	DbMaxOpenConns           int
	DbMaxIdleConns           int
	DbConnMaxLifetimeSeconds int
	DbConnMaxIdleTimeSeconds int
	DbMigrationsPath         string
)

// Deprecated: Use GlobalConfig.App...
var (
	ServerPort       string
	Issuer           string
	GroupAdminRoles  []string
	GroupUpdateRoles []string
	GroupAccessRoles []string
	AllowedOrigins   []string
)

// Deprecated: Use GlobalConfig.Minio...
var (
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioUseSSL    bool
	MinioBucket    string
)

// Deprecated: Use GlobalConfig.Redis...
var (
	RedisAddr           string
	RedisUsername       string
	RedisPassword       string
	RedisDB             int
	RedisUseTLS         bool
	RedisPoolSize       int
	RedisMinIdleConns   int
	RedisMaxRetries     int
	RedisDialTimeoutMs  int
	RedisReadTimeoutMs  int
	RedisWriteTimeoutMs int
	RedisPingTimeoutMs  int
	RedisAsyncQueue     int
	RedisAsyncWorkers   int
)

// Deprecated: Use GlobalConfig.K8s...
var (
	Scheme                     *runtime.Scheme
	DefaultStorageName         string
	DefaultStorageClassName    string
	DefaultStorageSize         string
	UserPVSize                 string
	ProjectPVSize              string
	UserStorageNs              string
	UserStoragePVC             string
	PersonalStorageServiceName string
	GroupStorageServiceName    string
	GroupStorageBrowserSVCName string
	HarborPrivatePrefix        string
)

// Deprecated: Use GlobalConfig.Environment...
var (
	IsProduction          bool
	ReservedGroupName     string
	ReservedAdminUsername string
)

// Deprecated: Use GlobalConfig.Scheduler...
var (
	ExecutorMode                       string
	FlashSchedEnabled                  bool
	SchedulerName                      string
	ConfigFilePriorityClassName        string
	ConfigFilePriorityValue            int
	ConfigFileQueueName                string
	ConfigFileJobQueueName             string
	DefaultQueueName                   string
	ConfigFileQueuePriority            int64
	ConfigFileJobQueuePriority         int64
	DefaultQueuePriority               int64
	ConfigFileQueuePreemptible         bool
	ConfigFileJobQueuePreemptible      bool
	DefaultQueuePreemptible            bool
	ConfigFileQueueMaxConcurrent       int64
	ConfigFileJobQueueMaxConcurrent    int64
	DefaultQueueMaxConcurrent          int64
	ConfigFileQueueTTLSeconds          int64
	ConfigFileJobQueueTTLSeconds       int64
	DefaultQueueTTLSeconds             int64
	FlashSchedQueueAnnotationKey       string
	FlashSchedPreemptableAnnotationKey string
	ConfigFilePreemptable              bool
)

// Deprecated: Use GlobalConfig.Prometheus...
var PrometheusAddr string

type Config struct {
	App         AppConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	Minio       MinioConfig
	K8s         K8sConfig
	Scheduler   SchedulerConfig
	Prometheus  PrometheusConfig
	Environment EnvironmentConfig
}

type AppConfig struct {
	JwtSecret        string
	ServerPort       string
	Issuer           string
	AllowedOrigins   []string
	GroupAdminRoles  []string
	GroupUpdateRoles []string
	GroupAccessRoles []string
}

type DatabaseConfig struct {
	Host                   string
	Port                   string
	User                   string
	Password               string
	Name                   string
	SSLMode                string
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeSeconds int
	ConnMaxIdleTimeSeconds int
	MigrationsPath         string
}

type RedisConfig struct {
	Addr           string
	Username       string
	Password       string
	DB             int
	UseTLS         bool
	PoolSize       int
	MinIdleConns   int
	MaxRetries     int
	DialTimeoutMs  int
	ReadTimeoutMs  int
	WriteTimeoutMs int
	PingTimeoutMs  int
	AsyncQueue     int
	AsyncWorkers   int
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

type K8sConfig struct {
	Scheme                     *runtime.Scheme
	DefaultStorageName         string
	DefaultStorageClassName    string
	DefaultStorageSize         string
	UserPVSize                 string
	ProjectPVSize              string
	UserStorageNs              string
	UserStoragePVC             string
	PersonalStorageServiceName string
	GroupStorageServiceName    string
	GroupStorageBrowserSVCName string
	HarborPrivatePrefix        string
}

type SchedulerConfig struct {
	ExecutorMode                       string
	FlashSchedEnabled                  bool
	SchedulerName                      string
	ConfigFilePriorityClassName        string
	ConfigFilePriorityValue            int
	ConfigFileQueueName                string
	ConfigFileJobQueueName             string
	DefaultQueueName                   string
	ConfigFileQueuePriority            int64
	ConfigFileJobQueuePriority         int64
	DefaultQueuePriority               int64
	ConfigFileQueuePreemptible         bool
	ConfigFileJobQueuePreemptible      bool
	DefaultQueuePreemptible            bool
	ConfigFileQueueMaxConcurrent       int64
	ConfigFileJobQueueMaxConcurrent    int64
	DefaultQueueMaxConcurrent          int64
	ConfigFileQueueTTLSeconds          int64
	ConfigFileJobQueueTTLSeconds       int64
	DefaultQueueTTLSeconds             int64
	FlashSchedQueueAnnotationKey       string
	FlashSchedPreemptableAnnotationKey string
	ConfigFilePreemptable              bool
}

type PrometheusConfig struct {
	Addr string
}

type EnvironmentConfig struct {
	IsProduction          bool
	ReservedGroupName     string
	ReservedAdminUsername string
}

// LoadConfig loads the configuration from environment variables.
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		App: AppConfig{
			JwtSecret:        getEnv("JWT_SECRET", "defaultsecret"),
			ServerPort:       getEnv("SERVER_PORT", "8080"),
			Issuer:           getEnv("ISSUER", "platform"),
			GroupAdminRoles:  []string{"admin"},
			GroupUpdateRoles: []string{"admin", "manager"},
			GroupAccessRoles: []string{"admin", "manager", "user"},
		},
		Database: DatabaseConfig{
			Host:           getEnv("DB_HOST", "postgres"),
			Port:           getEnv("DB_PORT", "5432"),
			User:           getEnv("DB_USER", "postgres"),
			Password:       getEnv("DB_PASSWORD", "password"),
			Name:           getEnv("DB_NAME", "platform"),
			SSLMode:        getEnv("DB_SSLMODE", "disable"),
			MigrationsPath: getEnv("DB_MIGRATIONS_PATH", "file://infra/db/migrations"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "redis:6379"),
			Username: getEnv("REDIS_USERNAME", ""),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		Minio: MinioConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "minio.tenant.svc.cluster.local:443"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minio"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minio123"),
			Bucket:    getEnv("MINIO_BUCKET", "platform-bucket"),
		},
		K8s: K8sConfig{
			Scheme:                     runtime.NewScheme(),
			DefaultStorageName:         "project",
			DefaultStorageClassName:    getEnv("DEFAULT_STORAGE_CLASS_NAME", "longhorn"),
			DefaultStorageSize:         "3Gi",
			UserPVSize:                 "10Gi",
			ProjectPVSize:              "10Gi",
			UserStorageNs:              "user-%s-storage",
			UserStoragePVC:             "user-%s-disk",
			PersonalStorageServiceName: getEnv("PERSONAL_STORAGE_SERVICE_NAME", "storage-svc"),
			GroupStorageServiceName:    getEnv("GROUP_STORAGE_SERVICE_NAME", "storage-svc"),
			GroupStorageBrowserSVCName: getEnv("GROUP_STORAGE_BROWSER_SVC_NAME", "filebrowser-group-svc"),
			HarborPrivatePrefix:        getEnv("HARBOR_PRIVATE_PREFIX", ""),
		},
		Scheduler: SchedulerConfig{
			ExecutorMode:                       getEnv("EXECUTOR_MODE", "local"),
			SchedulerName:                      getEnv("SCHEDULER_NAME", "flash-scheduler"),
			ConfigFilePriorityClassName:        getEnv("CONFIGFILE_PRIORITY_CLASS", "platform-configfile-critical"),
			ConfigFileQueueName:                getEnv("CONFIGFILE_QUEUE_NAME", "configfile-interactive"),
			ConfigFileJobQueueName:             getEnv("CONFIGFILE_JOB_QUEUE_NAME", "configfile-job"),
			DefaultQueueName:                   getEnv("DEFAULT_QUEUE_NAME", "default-batch"),
			FlashSchedQueueAnnotationKey:       getEnv("FLASH_SCHED_QUEUE_ANNOTATION_KEY", "scheduling.flash-sched.io/queue-name"),
			FlashSchedPreemptableAnnotationKey: getEnv("FLASH_SCHED_PREEMPTABLE_ANNOTATION_KEY", "scheduling.flash-sched.io/preemptable"),
		},
		Prometheus: PrometheusConfig{
			Addr: getEnv("PROMETHEUS_ADDR", "http://prometheus:9090"),
		},
		Environment: EnvironmentConfig{
			ReservedGroupName:     "super",
			ReservedAdminUsername: "admin",
		},
	}

	// Parsing with error logging
	cfg.Database.MaxOpenConns = parseEnvInt("DB_MAX_OPEN_CONNS", 25)
	cfg.Database.MaxIdleConns = parseEnvInt("DB_MAX_IDLE_CONNS", 10)
	cfg.Database.ConnMaxLifetimeSeconds = parseEnvInt("DB_CONN_MAX_LIFETIME_SECONDS", 300)
	cfg.Database.ConnMaxIdleTimeSeconds = parseEnvInt("DB_CONN_MAX_IDLE_TIME_SECONDS", 180)

	if origins := getEnv("ALLOWED_ORIGINS", ""); origins != "" {
		cfg.App.AllowedOrigins = strings.Split(origins, ",")
	}

	cfg.Redis.DB = parseEnvInt("REDIS_DB", 0)
	cfg.Redis.UseTLS = parseEnvBool("REDIS_USE_TLS", false)
	cfg.Redis.PoolSize = parseEnvInt("REDIS_POOL_SIZE", 10)
	cfg.Redis.MinIdleConns = parseEnvInt("REDIS_MIN_IDLE_CONNS", 2)
	cfg.Redis.MaxRetries = parseEnvInt("REDIS_MAX_RETRIES", 3)
	cfg.Redis.DialTimeoutMs = parseEnvInt("REDIS_DIAL_TIMEOUT_MS", 3000)
	cfg.Redis.ReadTimeoutMs = parseEnvInt("REDIS_READ_TIMEOUT_MS", 2000)
	cfg.Redis.WriteTimeoutMs = parseEnvInt("REDIS_WRITE_TIMEOUT_MS", 2000)
	cfg.Redis.PingTimeoutMs = parseEnvInt("REDIS_PING_TIMEOUT_MS", 1500)
	cfg.Redis.AsyncQueue = parseEnvInt("REDIS_ASYNC_QUEUE", 256)
	cfg.Redis.AsyncWorkers = parseEnvInt("REDIS_ASYNC_WORKERS", 2)

	cfg.Minio.UseSSL = parseEnvBool("MINIO_USE_SSL", true)

	env := getEnv("GO_ENV", "development")
	cfg.Environment.IsProduction = env == "production" || env == "release"

	cfg.Scheduler.FlashSchedEnabled = parseEnvBool("FLASH_SCHED_ENABLED", false)
	cfg.Scheduler.ConfigFilePriorityValue = parseEnvInt("CONFIGFILE_PRIORITY_VALUE", 1000000)
	cfg.Scheduler.ConfigFileQueuePriority = parseEnvInt64("CONFIGFILE_QUEUE_PRIORITY", 10000)
	cfg.Scheduler.ConfigFileJobQueuePriority = parseEnvInt64("CONFIGFILE_JOB_QUEUE_PRIORITY", 5000)
	cfg.Scheduler.DefaultQueuePriority = parseEnvInt64("DEFAULT_QUEUE_PRIORITY", 100)
	cfg.Scheduler.ConfigFileQueuePreemptible = parseEnvBool("CONFIGFILE_QUEUE_PREEMPTIBLE", false)
	cfg.Scheduler.ConfigFileJobQueuePreemptible = parseEnvBool("CONFIGFILE_JOB_QUEUE_PREEMPTIBLE", false)
	cfg.Scheduler.DefaultQueuePreemptible = parseEnvBool("DEFAULT_QUEUE_PREEMPTIBLE", true)
	cfg.Scheduler.ConfigFileQueueMaxConcurrent = parseEnvInt64("CONFIGFILE_QUEUE_MAX_CONCURRENT", 0)
	cfg.Scheduler.ConfigFileJobQueueMaxConcurrent = parseEnvInt64("CONFIGFILE_JOB_QUEUE_MAX_CONCURRENT", 0)
	cfg.Scheduler.DefaultQueueMaxConcurrent = parseEnvInt64("DEFAULT_QUEUE_MAX_CONCURRENT", 0)
	cfg.Scheduler.ConfigFileQueueTTLSeconds = parseEnvInt64("CONFIGFILE_QUEUE_TTL_SECONDS", 0)
	cfg.Scheduler.ConfigFileJobQueueTTLSeconds = parseEnvInt64("CONFIGFILE_JOB_QUEUE_TTL_SECONDS", 0)
	cfg.Scheduler.DefaultQueueTTLSeconds = parseEnvInt64("DEFAULT_QUEUE_TTL_SECONDS", 0)
	cfg.Scheduler.ConfigFilePreemptable = parseEnvBool("CONFIGFILE_PREEMPTABLE", false)

	// Init K8s Scheme
	initK8sScheme(cfg.K8s.Scheme)

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	// Populate global variables for backward compatibility
	GlobalConfig = cfg
	JwtSecret = cfg.App.JwtSecret
	ServerPort = cfg.App.ServerPort
	Issuer = cfg.App.Issuer
	AllowedOrigins = cfg.App.AllowedOrigins
	GroupAdminRoles = cfg.App.GroupAdminRoles
	GroupUpdateRoles = cfg.App.GroupUpdateRoles
	GroupAccessRoles = cfg.App.GroupAccessRoles

	DbHost = cfg.Database.Host
	DbPort = cfg.Database.Port
	DbUser = cfg.Database.User
	DbPassword = cfg.Database.Password
	DbName = cfg.Database.Name
	DbSSLMode = cfg.Database.SSLMode
	DbMaxOpenConns = cfg.Database.MaxOpenConns
	DbMaxIdleConns = cfg.Database.MaxIdleConns
	DbConnMaxLifetimeSeconds = cfg.Database.ConnMaxLifetimeSeconds
	DbConnMaxIdleTimeSeconds = cfg.Database.ConnMaxIdleTimeSeconds
	DbMigrationsPath = cfg.Database.MigrationsPath

	RedisAddr = cfg.Redis.Addr
	RedisUsername = cfg.Redis.Username
	RedisPassword = cfg.Redis.Password
	RedisDB = cfg.Redis.DB
	RedisUseTLS = cfg.Redis.UseTLS
	RedisPoolSize = cfg.Redis.PoolSize
	RedisMinIdleConns = cfg.Redis.MinIdleConns
	RedisMaxRetries = cfg.Redis.MaxRetries
	RedisDialTimeoutMs = cfg.Redis.DialTimeoutMs
	RedisReadTimeoutMs = cfg.Redis.ReadTimeoutMs
	RedisWriteTimeoutMs = cfg.Redis.WriteTimeoutMs
	RedisPingTimeoutMs = cfg.Redis.PingTimeoutMs
	RedisAsyncQueue = cfg.Redis.AsyncQueue
	RedisAsyncWorkers = cfg.Redis.AsyncWorkers

	MinioEndpoint = cfg.Minio.Endpoint
	MinioAccessKey = cfg.Minio.AccessKey
	MinioSecretKey = cfg.Minio.SecretKey
	MinioBucket = cfg.Minio.Bucket
	MinioUseSSL = cfg.Minio.UseSSL

	Scheme = cfg.K8s.Scheme
	DefaultStorageName = cfg.K8s.DefaultStorageName
	DefaultStorageClassName = cfg.K8s.DefaultStorageClassName
	DefaultStorageSize = cfg.K8s.DefaultStorageSize
	UserPVSize = cfg.K8s.UserPVSize
	ProjectPVSize = cfg.K8s.ProjectPVSize
	UserStorageNs = cfg.K8s.UserStorageNs
	UserStoragePVC = cfg.K8s.UserStoragePVC
	PersonalStorageServiceName = cfg.K8s.PersonalStorageServiceName
	GroupStorageServiceName = cfg.K8s.GroupStorageServiceName
	GroupStorageBrowserSVCName = cfg.K8s.GroupStorageBrowserSVCName
	HarborPrivatePrefix = cfg.K8s.HarborPrivatePrefix

	IsProduction = cfg.Environment.IsProduction
	ReservedGroupName = cfg.Environment.ReservedGroupName
	ReservedAdminUsername = cfg.Environment.ReservedAdminUsername

	ExecutorMode = cfg.Scheduler.ExecutorMode
	FlashSchedEnabled = cfg.Scheduler.FlashSchedEnabled
	SchedulerName = cfg.Scheduler.SchedulerName
	ConfigFilePriorityClassName = cfg.Scheduler.ConfigFilePriorityClassName
	ConfigFilePriorityValue = cfg.Scheduler.ConfigFilePriorityValue
	ConfigFileQueueName = cfg.Scheduler.ConfigFileQueueName
	ConfigFileJobQueueName = cfg.Scheduler.ConfigFileJobQueueName
	DefaultQueueName = cfg.Scheduler.DefaultQueueName
	ConfigFileQueuePriority = cfg.Scheduler.ConfigFileQueuePriority
	ConfigFileJobQueuePriority = cfg.Scheduler.ConfigFileJobQueuePriority
	DefaultQueuePriority = cfg.Scheduler.DefaultQueuePriority
	ConfigFileQueuePreemptible = cfg.Scheduler.ConfigFileQueuePreemptible
	ConfigFileJobQueuePreemptible = cfg.Scheduler.ConfigFileJobQueuePreemptible
	DefaultQueuePreemptible = cfg.Scheduler.DefaultQueuePreemptible
	ConfigFileQueueMaxConcurrent = cfg.Scheduler.ConfigFileQueueMaxConcurrent
	ConfigFileJobQueueMaxConcurrent = cfg.Scheduler.ConfigFileJobQueueMaxConcurrent
	DefaultQueueMaxConcurrent = cfg.Scheduler.DefaultQueueMaxConcurrent
	ConfigFileQueueTTLSeconds = cfg.Scheduler.ConfigFileQueueTTLSeconds
	ConfigFileJobQueueTTLSeconds = cfg.Scheduler.ConfigFileJobQueueTTLSeconds
	DefaultQueueTTLSeconds = cfg.Scheduler.DefaultQueueTTLSeconds
	FlashSchedQueueAnnotationKey = cfg.Scheduler.FlashSchedQueueAnnotationKey
	FlashSchedPreemptableAnnotationKey = cfg.Scheduler.FlashSchedPreemptableAnnotationKey
	ConfigFilePreemptable = cfg.Scheduler.ConfigFilePreemptable

	PrometheusAddr = cfg.Prometheus.Addr

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func parseEnvInt(key string, fallback int) int {
	raw := getEnv(key, "")
	if raw == "" {
		return fallback
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		log.Printf("WARNING: Invalid integer for %s: %v. Using default: %d", key, err, fallback)
		return fallback
	}
	return val
}

func parseEnvInt64(key string, fallback int64) int64 {
	raw := getEnv(key, "")
	if raw == "" {
		return fallback
	}
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		log.Printf("WARNING: Invalid integer64 for %s: %v. Using default: %d", key, err, fallback)
		return fallback
	}
	return val
}

func parseEnvBool(key string, fallback bool) bool {
	raw := getEnv(key, "")
	if raw == "" {
		return fallback
	}
	val, err := strconv.ParseBool(raw)
	if err != nil {
		log.Printf("WARNING: Invalid boolean for %s: %v. Using default: %v", key, err, fallback)
		return fallback
	}
	return val
}

func validateConfig(cfg *Config) error {
	insecureDefaults := map[string]string{
		"JWT_SECRET":       "defaultsecret",
		"DB_PASSWORD":      "password",
		"MINIO_SECRET_KEY": "minio123",
	}

	vals := map[string]string{
		"JWT_SECRET":       cfg.App.JwtSecret,
		"DB_PASSWORD":      cfg.Database.Password,
		"MINIO_SECRET_KEY": cfg.Minio.SecretKey,
	}

	for envKey, currentVal := range vals {
		if def, ok := insecureDefaults[envKey]; ok && currentVal == def {
			if cfg.Environment.IsProduction {
				return fmt.Errorf("SECURITY: %s is using insecure default value in production", envKey)
			}
			log.Printf("WARNING: %s is using insecure default value '%s'", envKey, def)
		}
	}

	if cfg.K8s.HarborPrivatePrefix == "" {
		if cfg.Environment.IsProduction {
			return fmt.Errorf("SECURITY: HARBOR_PRIVATE_PREFIX must be set in production")
		}
		log.Println("WARNING: HARBOR_PRIVATE_PREFIX is not set")
	}

	if cfg.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS must be > 0")
	}
	if cfg.Redis.PoolSize <= 0 {
		return fmt.Errorf("REDIS_POOL_SIZE must be > 0")
	}

	return nil
}

func initK8sScheme(scheme *runtime.Scheme) {
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
}
