package source

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitpatrol/internal/models"
)

var gitlabRateLimitUntil time.Time

type GitLabSource struct {
	token string
}

func (s *GitLabSource) get(apiURL string) (*http.Response, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	if s.token != "" {
		req.Header.Set("PRIVATE-TOKEN", s.token)
	}
	return http.DefaultClient.Do(req)
}

func (s *GitLabSource) encodedPath(repoURL string) string {
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	namespace := parts[len(parts)-2]
	project := parts[len(parts)-1]
	return url.PathEscape(namespace + "/" + project)
}

func (s *GitLabSource) GetMetadata(repoURL string) (models.Metadata, error) {
	if time.Now().Before(gitlabRateLimitUntil) {
		return models.Metadata{}, fmt.Errorf("GitLab API rate limit active, skipping metadata")
	}

	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	if len(parts) < 2 {
		return models.Metadata{}, fmt.Errorf("invalid URL")
	}

	resp, err := s.get(fmt.Sprintf("https://gitlab.com/api/v4/projects/%s", s.encodedPath(repoURL)))
	if err != nil {
		return models.Metadata{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		gitlabRateLimitUntil = time.Now().Add(15 * time.Minute)
		log.Printf("[SOURCE] GitLab Rate Limit hit. Pausing API calls for 15 minutes.")
		return models.Metadata{}, fmt.Errorf("GitLab API rate limit hit")
	}

	if resp.StatusCode != 200 {
		return models.Metadata{}, fmt.Errorf("GitLab API returned %d", resp.StatusCode)
	}

	var meta struct {
		StarCount       int    `json:"star_count"`
		ForksCount      int    `json:"forks_count"`
		OpenIssuesCount int    `json:"open_issues_count"`
		AvatarURL       string `json:"avatar_url"`
		Namespace       struct {
			Path string `json:"path"`
		} `json:"namespace"`
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &meta)

	return models.Metadata{
		Stars:      meta.StarCount,
		Forks:      meta.ForksCount,
		OpenIssues: meta.OpenIssuesCount,
		AvatarURL:  meta.AvatarURL,
		Username:   meta.Namespace.Path,
	}, nil
}

func (s *GitLabSource) GetWikiURL(repoURL string) (string, bool) {
	wikiURL := strings.TrimSuffix(repoURL, ".git") + ".wiki.git"
	return wikiURL, true
}

func (s *GitLabSource) SyncIssues(repoURL string, destPath string) error {
	if time.Now().Before(gitlabRateLimitUntil) {
		return nil
	}

	resp, err := s.get(fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/issues?state=all&per_page=100", s.encodedPath(repoURL)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		gitlabRateLimitUntil = time.Now().Add(15 * time.Minute)
		return fmt.Errorf("GitLab API rate limit hit")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitLab API returned %d", resp.StatusCode)
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

func (s *GitLabSource) SyncReleases(repoURL string, destPath string) error {
	if time.Now().Before(gitlabRateLimitUntil) {
		return nil
	}

	resp, err := s.get(fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/releases?per_page=100", s.encodedPath(repoURL)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		gitlabRateLimitUntil = time.Now().Add(15 * time.Minute)
		return fmt.Errorf("GitLab API rate limit hit")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitLab API returned %d", resp.StatusCode)
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
