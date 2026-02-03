package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/repository"
)

// TODO: Refactor Handler structure - implement sub-packages: image, k8s, project, user
// These handlers are currently in the same package and need to be reorganized

type Handlers struct {
	Audit       *AuditHandler
	ConfigFile  *ConfigFileHandler
	Group       *GroupHandler
	Form        *FormHandler
	PVCBinding  *PVCBindingHandler
	FileBrowser *FileBrowserHandler
	Router      *gin.Engine
}

func New(svc *application.Services, repos *repository.Repos, router *gin.Engine) *Handlers {
	h := &Handlers{
		Audit:       NewAuditHandler(svc.Audit),
		ConfigFile:  NewConfigFileHandler(svc.ConfigFile),
		Group:       NewGroupHandler(svc.Group),
		Form:        NewFormHandler(svc.Form),
		PVCBinding:  NewPVCBindingHandler(svc.K8s.PVCBindingManager),
		FileBrowser: NewFileBrowserHandler(svc.K8s.FileBrowserManager),
		Router:      router,
	}
	return h
}
