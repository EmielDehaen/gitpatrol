package source

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitpatrol/internal/models"
)

var githubRateLimitUntil time.Time

type Source interface {
	GetMetadata(url string) (models.Metadata, error)
	GetWikiURL(url string) (string, bool)
	SyncIssues(url string, destPath string) error
	SyncReleases(url string, destPath string) error
}

func GetSource(url string) (Source, error) {
	if strings.Contains(url, "github.com") {
		return &GitHubSource{}, nil
	}
	return nil, fmt.Errorf("provider not supported for %s", url)
}

type GitHubSource struct{}

func (s *GitHubSource) GetMetadata(url string) (models.Metadata, error) {
	if time.Now().Before(githubRateLimitUntil) {
		return models.Metadata{}, fmt.Errorf("GitHub API rate limit active, skipping metadata")
	}

	parts := strings.Split(strings.TrimSuffix(url, ".git"), "/")
	if len(parts) < 2 {
		return models.Metadata{}, fmt.Errorf("invalid URL")
	}
	repoPath := parts[len(parts)-2] + "/" + parts[len(parts)-1]
	username := parts[len(parts)-2]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s", repoPath)
	resp, err := http.Get(apiURL)
	if err != nil {
		return models.Metadata{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		githubRateLimitUntil = time.Now().Add(15 * time.Minute)
		log.Printf("[SOURCE] GitHub Rate Limit hit. Pausing API calls for 15 minutes.")
		return models.Metadata{}, fmt.Errorf("GitHub API rate limit hit")
	}

	if resp.StatusCode != 200 {
		return models.Metadata{}, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var meta struct {
		StargazersCount int `json:"stargazers_count"`
		ForksCount      int `json:"forks_count"`
		OpenIssuesCount int `json:"open_issues_count"`
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &meta)

	return models.Metadata{
		Stars:      meta.StargazersCount,
		Forks:      meta.ForksCount,
		OpenIssues: meta.OpenIssuesCount,
		AvatarURL:  fmt.Sprintf("https://github.com/%s.png?size=100", username),
		Username:   username,
	}, nil
}

func (s *GitHubSource) GetWikiURL(url string) (string, bool) {
	wikiURL := strings.TrimSuffix(url, ".git") + ".wiki.git"
	return wikiURL, true
}

func (s *GitHubSource) SyncIssues(url string, destPath string) error {
	if time.Now().Before(githubRateLimitUntil) {
		return nil
	}

	parts := strings.Split(strings.TrimSuffix(url, ".git"), "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid URL")
	}
	repoPath := parts[len(parts)-2] + "/" + parts[len(parts)-1]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/issues?state=all&per_page=100", repoPath)
	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		githubRateLimitUntil = time.Now().Add(15 * time.Minute)
		return fmt.Errorf("GitHub API rate limit hit")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	os.MkdirAll(destPath, 0755)
	outFile, err := os.Create(filepath.Join(destPath, "issues.json"))
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	return err
}

func (s *GitHubSource) SyncReleases(url string, destPath string) error {
	if time.Now().Before(githubRateLimitUntil) {
		return nil
	}

	parts := strings.Split(strings.TrimSuffix(url, ".git"), "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid URL")
	}
	repoPath := parts[len(parts)-2] + "/" + parts[len(parts)-1]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=100", repoPath)
	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		githubRateLimitUntil = time.Now().Add(15 * time.Minute)
		return fmt.Errorf("GitHub API rate limit hit")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	os.MkdirAll(destPath, 0755)
	outFile, err := os.Create(filepath.Join(destPath, "releases.json"))
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	return err
}
