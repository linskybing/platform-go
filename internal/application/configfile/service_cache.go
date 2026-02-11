package configfile

import (
	"context"
	"fmt"
)

func configFileListKey() string {
	return "cache:configfile:list"
}

func configFileByIDKey(id string) string {
	return fmt.Sprintf("cache:configfile:by-id:%s", id)
}

func configFileByProjectKey(projectID string) string {
	return fmt.Sprintf("cache:configfile:by-project:%s", projectID)
}

func (s *ConfigFileService) invalidateConfigFileCache(id, projectID string) {
	if s.cache == nil || !s.cache.Enabled() {
		return
	}
	ctx := context.Background()
	_ = s.cache.Invalidate(ctx, configFileListKey(), configFileByIDKey(id), configFileByProjectKey(projectID))
}
