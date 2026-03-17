package source

import (
	"fmt"
	"strings"

	"gitpatrol/internal/models"
)

type Source interface {
	GetMetadata(url string) (models.Metadata, error)
	GetWikiURL(url string) (string, bool)
	SyncIssues(url string, destPath string) error
	SyncReleases(url string, destPath string) error
}

func GetSource(url, githubToken, gitlabToken string) (Source, error) {
	if strings.Contains(url, "github.com") {
		return &GitHubSource{token: githubToken}, nil
	}
	if strings.Contains(url, "gitlab.com") {
		return &GitLabSource{token: gitlabToken}, nil
	}
	return nil, fmt.Errorf("provider not supported for %s", url)
}
