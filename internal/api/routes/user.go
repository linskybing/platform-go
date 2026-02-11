package routes

		// Admin User Management (usually) - access control checked inside handler or middleware
		// Register CRUD and listing endpoints.
		users.GET("/", h.User.GetUsers)
		users.GET("/paging", h.User.ListUsersPaging)
		users.GET(":id", h.User.GetUserByID)
	"github.com/linskybing/platform-go/internal/api/middleware"
)

func registerUserRoutes(r *gin.RouterGroup, h *handlers.Handlers, authMw *middleware.AuthMiddleware) {
	users := r.Group("/users")
	{
		// Admin User Management (usually) - access control checked inside handler or middleware
		// For now registering basic CRUD.
		users.GET("/:id", h.User.GetUserByID)
		users.PUT("/:id", h.User.UpdateUser)
		users.DELETE("/:id", h.User.DeleteUser)
		// users.GET("/", h.User.ListUsers) // If implemented
	}
}
