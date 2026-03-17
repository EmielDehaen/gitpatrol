package destination

import (
	"gitpatrol/internal/models"
)

type Destination interface {
	CreateRepository(name string, description string) (string, error)
	PushMirror(localPath string, targetURL string) error
	SyncMetadata(repo *models.Repository, targetURL string) error
}
