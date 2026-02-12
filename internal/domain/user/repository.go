package user

// Repository defines data access interface for users
type Repository interface {
	GetAllUsers() ([]UserWithSuperAdmin, error)
	ListUsersPaging(page, limit int) ([]UserWithSuperAdmin, error)
	GetUserByID(id string) (UserWithSuperAdmin, error)
	GetUsernameByID(id string) (string, error)
	GetUserByUsername(username string) (User, error)
	GetUserRawByID(id string) (User, error)
	SaveUser(user *User) error
	DeleteUser(id string) error
}
