package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	JwtSecret        string
	DbHost           string
	DbPort           string
	DbUser           string
	DbPassword       string
	DbName           string
	ServerPort       string
	Issuer           string
	GroupAdminRoles  = []string{"admin"}
	GroupUpdateRoles = []string{"admin", "manager"}
	GroupAccessRoles = []string{"admin", "manager", "user"}
	MinioEndpoint    string
	MinioAccessKey   string
	MinioSecretKey   string
	MinioUseSSL      bool
	MinioBucket      string
	// Redis
	RedisAddr               string
	RedisUsername           string
	RedisPassword           string
	RedisDB                 int
	RedisUseTLS             bool
	RedisPoolSize           int
	RedisMinIdleConns       int
	RedisMaxRetries         int
	RedisDialTimeoutMs      int
	RedisReadTimeoutMs      int
	RedisWriteTimeoutMs     int
	RedisPingTimeoutMs      int
	RedisAsyncQueue         int
	RedisAsyncWorkers       int
	Scheme                  = runtime.NewScheme()
	DefaultStorageName      = "project"
	DefaultStorageClassName = "longhorn"
	DefaultStorageSize      = "3Gi"
	UserPVSize              = "10Gi"
	ProjectPVSize           = "10Gi"
	// Environment
	IsProduction bool
	// Reserved names that cannot be deleted or downgraded
	ReservedGroupName     = "super"
	ReservedAdminUsername = "admin"
	// Storage Pattern
	UserStorageNs  = "user-%s-storage" // user-{username}-storage
	UserStoragePVC = "user-%s-disk"    // user-{username}-disk
	// K8s Service Names
	PersonalStorageServiceName string
	GroupStorageServiceName    string
	GroupStorageBrowserSVCName string
	HarborPrivatePrefix        string
)

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	JwtSecret = getEnv("JWT_SECRET", "defaultsecret")
	DbHost = getEnv("DB_HOST", "localhost")
	DbPort = getEnv("DB_PORT", "5432")
	DbUser = getEnv("DB_USER", "postgres")
	DbPassword = getEnv("DB_PASSWORD", "password")
	DbName = getEnv("DB_NAME", "platform")
	ServerPort = getEnv("SERVER_PORT", "8080")
	Issuer = getEnv("Issuer", "platform")

	MinioEndpoint = getEnv("MINIO_ENDPOINT", "minio.tenant.svc.cluster.local:443")
	MinioAccessKey = getEnv("MINIO_ACCESS_KEY", "minio")
	MinioSecretKey = getEnv("MINIO_SECRET_KEY", "minio123")
	MinioBucket = getEnv("MINIO_BUCKET", "platform-bucket")
	MinioUseSSL, _ = strconv.ParseBool(getEnv("MINIO_USE_SSL", "true"))

	RedisAddr = getEnv("REDIS_ADDR", "")
	RedisUsername = getEnv("REDIS_USERNAME", "")
	RedisPassword = getEnv("REDIS_PASSWORD", "")
	RedisDB, _ = strconv.Atoi(getEnv("REDIS_DB", "0"))
	RedisUseTLS, _ = strconv.ParseBool(getEnv("REDIS_USE_TLS", "false"))
	RedisPoolSize, _ = strconv.Atoi(getEnv("REDIS_POOL_SIZE", "10"))
	RedisMinIdleConns, _ = strconv.Atoi(getEnv("REDIS_MIN_IDLE_CONNS", "2"))
	RedisMaxRetries, _ = strconv.Atoi(getEnv("REDIS_MAX_RETRIES", "3"))
	RedisDialTimeoutMs, _ = strconv.Atoi(getEnv("REDIS_DIAL_TIMEOUT_MS", "3000"))
	RedisReadTimeoutMs, _ = strconv.Atoi(getEnv("REDIS_READ_TIMEOUT_MS", "2000"))
	RedisWriteTimeoutMs, _ = strconv.Atoi(getEnv("REDIS_WRITE_TIMEOUT_MS", "2000"))
	RedisPingTimeoutMs, _ = strconv.Atoi(getEnv("REDIS_PING_TIMEOUT_MS", "1500"))
	RedisAsyncQueue, _ = strconv.Atoi(getEnv("REDIS_ASYNC_QUEUE", "256"))
	RedisAsyncWorkers, _ = strconv.Atoi(getEnv("REDIS_ASYNC_WORKERS", "2"))

	DefaultStorageClassName = getEnv("DEFAULT_STORAGE_CLASS_NAME", "longhorn")

	// Environment
	env := getEnv("GO_ENV", "development")
	IsProduction = env == "production" || env == "release"

	// K8s Service Names
	PersonalStorageServiceName = getEnv("PERSONAL_STORAGE_SERVICE_NAME", "storage-svc")
	GroupStorageServiceName = getEnv("GROUP_STORAGE_SERVICE_NAME", "storage-svc")
	GroupStorageBrowserSVCName = getEnv("GROUP_STORAGE_BROWSER_SVC_NAME", "filebrowser-group-svc")
	HarborPrivatePrefix = getEnv("HARBOR_PRIVATE_PREFIX", "192.168.110.1:30003/library/")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func InitK8sConfig() {
	// Register core Kubernetes API types
	_ = corev1.AddToScheme(Scheme)
	_ = appsv1.AddToScheme(Scheme)
	_ = batchv1.AddToScheme(Scheme)
	_ = networkingv1.AddToScheme(Scheme)
}
