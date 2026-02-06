package group

type GroupUpdateDTO struct {
	GroupName   *string `json:"group_name" form:"group_name"`
	Description *string `json:"description" form:"description"`
}

type GroupCreateDTO struct {
	GroupName   string  `json:"group_name" form:"group_name" binding:"required"`
	Description *string `json:"description" form:"description"`
}

type UserGroupInputDTO struct {
	UID  string `json:"uid" form:"u_id" binding:"required"`
	GID  string `json:"gid" form:"g_id" binding:"required"`
	Role string `json:"role" form:"role" binding:"required,oneof=admin manager user"`
}

type UserGroupCreateDTO struct {
	UID  string `json:"uid" form:"u_id" binding:"required"`
	GID  string `json:"gid" form:"g_id" binding:"required"`
	Role string `json:"role" form:"role" binding:"oneof=admin manager user"`
}

type UserGroupRoleDTO struct {
	UID  string `json:"uid" form:"u_id" binding:"required"`
	GID  string `json:"gid" form:"g_id" binding:"required"`
	Role string `json:"role" form:"role" binding:"required,oneof=admin manager user"`
}

type UserGroupDeleteDTO struct {
	UID string `json:"uid" form:"u_id" binding:"required"`
	GID string `json:"gid" form:"g_id" binding:"required"`
}

func (d UserGroupInputDTO) GetGID() string {
	return d.GID
}

func (d UserGroupDeleteDTO) GetGID() string {
	return d.GID
}
