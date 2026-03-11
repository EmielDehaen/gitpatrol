package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Metadata struct {
	Stars      int
	Forks      int
	OpenIssues int
	AvatarURL  string
	Username   string
}

type Source interface {
	GetMetadata(url string) (Metadata, error)
	GetWikiURL(url string) (string, bool)
	SyncIssues(url string, destPath string) error
	SyncReleases(url string, destPath string) error
}

type GitHubSource struct{}

func (s *GitHubSource) GetMetadata(url string) (Metadata, error) {
	parts := strings.Split(strings.TrimSuffix(url, ".git"), "/")
	if len(parts) < 2 {
		return Metadata{}, fmt.Errorf("invalid URL")
	}
	repoPath := parts[len(parts)-2] + "/" + parts[len(parts)-1]
	username := parts[len(parts)-2]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s", repoPath)
	resp, err := http.Get(apiURL)
	if err != nil {
		return Metadata{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Metadata{}, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var meta struct {
		StargazersCount int `json:"stargazers_count"`
		ForksCount      int `json:"forks_count"`
		OpenIssuesCount int `json:"open_issues_count"`
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &meta)

	return Metadata{
		Stars:      meta.StargazersCount,
		Forks:      meta.ForksCount,
		OpenIssues: meta.OpenIssuesCount,
		AvatarURL:  fmt.Sprintf("https://github.com/%s.png?size=100", username),
		Username:   username,
	}, nil
}

func (s *GitHubSource) GetWikiURL(url string) (string, bool) {
	// GitHub Wikis are always at .wiki.git
	wikiURL := strings.TrimSuffix(url, ".git") + ".wiki.git"
	return wikiURL, true
}

func (s *GitHubSource) SyncIssues(url string, destPath string) error {
	parts := strings.Split(strings.TrimSuffix(url, ".git"), "/")
	if len(parts) < 2 { return fmt.Errorf("invalid URL") }
	repoPath := parts[len(parts)-2] + "/" + parts[len(parts)-1]
	
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/issues?state=all&per_page=100", repoPath)
	resp, err := http.Get(apiURL)
	if err != nil { return err }
	defer resp.Body.Close()

	if resp.StatusCode != 200 { return fmt.Errorf("GitHub API returned %d", resp.StatusCode) }

	os.MkdirAll(destPath, 0755)
	outFile, err := os.Create(filepath.Join(destPath, "issues.json"))
	if err != nil { return err }
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	return err
}

func (s *GitHubSource) SyncReleases(url string, destPath string) error {
	parts := strings.Split(strings.TrimSuffix(url, ".git"), "/")
	if len(parts) < 2 { return fmt.Errorf("invalid URL") }
	repoPath := parts[len(parts)-2] + "/" + parts[len(parts)-1]
	
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=100", repoPath)
	resp, err := http.Get(apiURL)
	if err != nil { return err }
	defer resp.Body.Close()

	if resp.StatusCode != 200 { return fmt.Errorf("GitHub API returned %d", resp.StatusCode) }

	os.MkdirAll(destPath, 0755)
	outFile, err := os.Create(filepath.Join(destPath, "releases.json"))
	if err != nil { return err }
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	return err
}

type GitLabSource struct{}

func (s *GitLabSource) GetMetadata(url string) (Metadata, error) {
	// Simple stub for now
	return Metadata{
		Stars:      0,
		Forks:      0,
		OpenIssues: 0,
		AvatarURL:  "",
		Username:   "",
	}, nil
}

func (s *GitLabSource) GetWikiURL(url string) (string, bool) { return "", false }
func (s *GitLabSource) SyncIssues(url string, destPath string) error { return nil }
func (s *GitLabSource) SyncReleases(url string, destPath string) error { return nil }

func GetSource(url string) (Source, error) {
	if strings.Contains(url, "github.com") {
		return &GitHubSource{}, nil
	}
	if strings.Contains(url, "gitlab.com") {
		return &GitLabSource{}, nil
	}
	return nil, fmt.Errorf("unknown source for URL: %s", url)
}

func downloadAvatar(url string, username string) error {
	if url == "" {
		return nil
	}
	avatarPath := filepath.Join("./data/avatars", username+".png")
	os.MkdirAll("./data/avatars", 0755)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(avatarPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
