package view

type ProjectGroupView struct {
	GID           string `gorm:"column:g_id" json:"GID"`
	GroupName     string `gorm:"column:group_name" json:"GroupName"`
	ProjectCount  int64  `gorm:"column:project_count" json:"ProjectCount"`
	ResourceCount int64  `gorm:"column:resource_count" json:"ResourceCount"`
	GroupCreateAt string `gorm:"column:group_create_at" json:"GroupCreateAt"`
	GroupUpdateAt string `gorm:"column:group_update_at" json:"GroupUpdateAt"`
}

type ProjectResourceView struct {
	PID              string `gorm:"column:p_id" json:"PID"`
	ProjectName      string `gorm:"column:project_name" json:"ProjectName"`
	RID              string `gorm:"column:r_id" json:"RID"`
	Type             string `gorm:"column:type" json:"Type"`
	Name             string `gorm:"column:name" json:"Name"`
	Filename         string `gorm:"column:filename" json:"Filename"`
	ResourceCreateAt string `gorm:"column:resource_create_at" json:"ResourceCreateAt"`
	ResourceUpdateAt string `gorm:"column:resource_update_at" json:"ResourceUpdateAt"`
}

type GroupResourceView struct {
	GID              string `gorm:"column:g_id" json:"GID"`
	GroupName        string `gorm:"column:group_name" json:"GroupName"`
	PID              string `gorm:"column:p_id" json:"PID"`
	ProjectName      string `gorm:"column:project_name" json:"ProjectName"`
	RID              string `gorm:"column:r_id" json:"RID"`
	ResourceType     string `gorm:"column:resource_type" json:"ResourceType"`
	ResourceName     string `gorm:"column:resource_name" json:"ResourceName"`
	Filename         string `gorm:"column:filename" json:"Filename"`
	ResourceCreateAt string `gorm:"column:resource_create_at" json:"ResourceCreateAt"`
	ResourceUpdateAt string `gorm:"column:resource_update_at" json:"ResourceUpdateAt"`
}

type UserGroupView struct {
	UID       string `gorm:"column:u_id" json:"UID"`
	Username  string `gorm:"column:username" json:"Username"`
	GID       string `gorm:"column:g_id" json:"GID"`
	GroupName string `gorm:"column:group_name" json:"GroupName"`
	Role      string `gorm:"column:role" json:"Role"`
}

type ProjectUserView struct {
	PID         string `gorm:"column:p_id" json:"PID"`
	ProjectName string `gorm:"column:project_name" json:"ProjectName"`
	GID         string `gorm:"column:g_id" json:"GID"`
	GroupName   string `gorm:"column:group_name" json:"GroupName"`
	Role        string `gorm:"column:role" json:"Role"`
	UID         string `gorm:"column:u_id" json:"UID"`
	Username    string `gorm:"column:username" json:"Username"`
}
