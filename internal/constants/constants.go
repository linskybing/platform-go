package constants

// Database view names - centralized to avoid magic strings
const (
	ViewProjectGroup    = "project_group_views"
	ViewProjectResource = "project_resource_views"
	ViewGroupResource   = "group_resource_views"
	ViewUsersSuperadmin = "users_with_superadmin"
	ViewUserGroup       = "user_group_views"
	ViewProjectUser     = "project_user_views"
)

// K8s namespace and resource naming patterns
const (
	// User storage namespace pattern
	UserStorageNamespacePattern = "user-%s-storage"
	UserStoragePVCPattern       = "user-%s-disk"

	// Group storage namespace pattern
	GroupStorageNamespacePattern = "group-%d-storage"
	GroupPVCIDPattern            = "group-%d-%s"
	GroupPVCNamePattern          = "group-%d-disk"

	// UUID short length for resource naming
	UUIDShortLength = 8
)

// Context timeout durations
const (
	// K8s operation timeouts
	K8sQuickOpTimeout    = 10  // seconds - for quick K8s operations
	K8sStandardOpTimeout = 30  // seconds - for standard K8s operations
	K8sImagePullTimeout  = 1200 // seconds (20 minutes) - for image pull operations

	// Database operation timeouts
	DBQueryTimeout       = 10 // seconds - for database queries
	DBTransactionTimeout = 30 // seconds - for database transactions
)

// Cache configuration
const (
	// PVC cache TTL in minutes
	PVCCacheTTL = 5

	// Maximum concurrent K8s API calls
	MaxConcurrentK8sCalls = 3
)

// Password and security constants
const (
	// Bcrypt cost for password hashing (12 is production-grade)
	BcryptCost = 12

	// Minimum password length
	MinPasswordLength = 8

	// JWT token expiry in hours
	JWTExpiryHours = 24
)

// HTTP status messages
const (
	MsgSuccess       = "Operation successful"
	MsgCreated       = "Resource created successfully"
	MsgUpdated       = "Resource updated successfully"
	MsgDeleted       = "Resource deleted successfully"
	MsgNotFound      = "Resource not found"
	MsgUnauthorized  = "Unauthorized"
	MsgForbidden     = "Access denied"
	MsgBadRequest    = "Invalid request"
	MsgInternalError = "Internal server error"
)

// Password security constants
const (
	// Minimum lengths
	MinUsernameLength = 3
	MaxUsernameLength = 50
)

// Resource limits
const (
	// Default storage size in Gi
	DefaultStorageSizeGi = 10

	// Maximum file upload size in MB
	MaxFileUploadSizeMB = 100

	// API pagination defaults
	DefaultPageSize = 20
	MaxPageSize     = 100
)
