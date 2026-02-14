package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/types"
)

// AuthMiddleware handles role-based authorization for API endpoints.
type AuthMiddleware struct {
	repos *repository.Repos
}

// NewAuthMiddleware creates a new authorization middleware instance.
func NewAuthMiddleware(repos *repository.Repos) *AuthMiddleware {
	return &AuthMiddleware{repos: repos}
}

// Admin restricts endpoint access to super admin only.
// Use for: System configuration, global operations, admin panels.
func (am *AuthMiddleware) Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		userClaims := claims.(*types.Claims)
		if !userClaims.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GroupAdmin restricts access to super admin or group admin for specific group.
// Use for: Group management, project deletion, group-level operations.
func (am *AuthMiddleware) GroupAdmin(extractor IDExtractor) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		userClaims := claims.(*types.Claims)

		// Super admin bypasses all checks
		if userClaims.IsAdmin {
			c.Next()
			return
		}

		// Extract group ID from request
		groupID := extractor(c, am.repos)
		if groupID == "" {
			// Extractor already set error response
			c.Abort()
			return
		}

		// Check if user is admin of this group
		userGroup, err := am.repos.UserGroup.GetUserGroup(ctx, userClaims.UserID, groupID)
		if err != nil || userGroup.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "group admin access required",
				"required": "group_admin",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GroupManager restricts access to super admin, group admin, or group manager.
// Use for: Project updates, resource configuration, management operations.
func (am *AuthMiddleware) GroupManager(extractor IDExtractor) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		userClaims := claims.(*types.Claims)

		// Super admin bypasses all checks
		if userClaims.IsAdmin {
			c.Next()
			return
		}

		// Extract group ID from request
		groupID := extractor(c, am.repos)
		if groupID == "" {
			// Extractor already set error response
			c.Abort()
			return
		}

		// Check if user is admin or manager of this group
		userGroup, err := am.repos.UserGroup.GetUserGroup(ctx, userClaims.UserID, groupID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "group manager access required",
				"required": "group_manager",
			})
			c.Abort()
			return
		}

		if userGroup.Role != "admin" && userGroup.Role != "manager" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "group manager access required",
				"required": "group_manager",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GroupMember restricts access to group members (including admin and manager).
// Use for: User submissions, viewing resources, member-level operations.
func (am *AuthMiddleware) GroupMember(extractor IDExtractor) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		userClaims := claims.(*types.Claims)

		// Super admin bypasses all checks
		if userClaims.IsAdmin {
			c.Next()
			return
		}

		// Extract group ID from request
		groupID := extractor(c, am.repos)
		if groupID == "" {
			// Extractor already set error response
			c.Abort()
			return
		}

		// Check if user is member of this group (any role)
		_, err := am.repos.UserGroup.GetUserGroup(ctx, userClaims.UserID, groupID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "group membership required",
				"required": "group_member",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserOrAdmin restricts access to the user themselves or super admin.
// Use for: Profile updates, account management, self-service operations.
func (am *AuthMiddleware) UserOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		userClaims := claims.(*types.Claims)

		// Super admin bypasses all checks
		if userClaims.IsAdmin {
			c.Next()
			return
		}

		// Extract user ID from URL parameter
		userID := c.Param("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			c.Abort()
			return
		}

		// Check if user is accessing their own resource
		if userClaims.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "can only access own resources",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
