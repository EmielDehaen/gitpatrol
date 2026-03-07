package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type GitHubMeta struct {
	StargazersCount int `json:"stargazers_count"`
	ForksCount      int `json:"forks_count"`
	OpenIssuesCount int `json:"open_issues_count"`
}

func fetchGitHubMeta(url string) (GitHubMeta, error) {
	parts := strings.Split(strings.TrimSuffix(url, ".git"), "/")
	if len(parts) < 2 {
		return GitHubMeta{}, fmt.Errorf("invalid URL")
	}
	repoPath := parts[len(parts)-2] + "/" + parts[len(parts)-1]
	
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s", repoPath)
	resp, err := http.Get(apiURL)
	if err != nil {
		return GitHubMeta{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return GitHubMeta{}, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var meta GitHubMeta
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &meta)
	return meta, nil
}

func downloadAvatar(url string, username string) error {
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

func syncRepo(id int, url, name string) {
	updateStatus(id, "syncing", "")
	
	parts := strings.Split(strings.TrimSuffix(url, ".git"), "/")
	username := ""
	if len(parts) >= 2 {
		username = parts[len(parts)-2]
	}

	repoPath := filepath.Join("./data", name)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		// Normal clone
		cmd := exec.Command("git", "clone", url, repoPath)
		if err := cmd.Run(); err != nil {
			updateStatus(id, "error", err.Error())
			return
		}
	} else {
		// Accumulative archive fetch: keep everything, even if deleted on remote
		cmd := exec.Command("git", "-C", repoPath, "fetch", "--all", "--tags", "--force")
		if err := cmd.Run(); err != nil {
			updateStatus(id, "error", err.Error())
			return
		}
	}

	meta, _ := fetchGitHubMeta(url)
	if username != "" {
		avatarURL := fmt.Sprintf("https://github.com/%s.png?size=100", username)
		downloadAvatar(avatarURL, username)
	}

	history := getCommitHistory(repoPath)
	lastCommits := getLastCommits(repoPath)
	
	cmd := exec.Command("git", "-C", repoPath, "log", "-1", "--format=%cI")
	output, _ := cmd.CombinedOutput()
	lastCommitTime, _ := time.Parse(time.RFC3339, strings.TrimSpace(string(output)))
	
	score := calculateHealthScore(meta, history, lastCommitTime)

	// Get default branch name
	branchCmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, _ := branchCmd.CombinedOutput()
	defaultBranch := strings.TrimSpace(string(branchOut))
	if defaultBranch == "" { defaultBranch = "main" }

	db.Exec(`UPDATE repositories SET 
		status = 'synced', 
		last_sync = ?, 
		last_commit = ?, 
		stars = ?, 
		forks = ?, 
		open_issues = ?, 
		commit_history = ?, 
		health_score = ?,
		default_branch = ?,
		error_message = '' 
		WHERE id = ?`, 
		time.Now(), lastCommits, meta.StargazersCount, meta.ForksCount, meta.OpenIssuesCount, history, score, defaultBranch, id)
	
	broadcastStatus(id, "synced", "")
}

func getCommitHistory(repoPath string) string {
	history := make([]int, 14)
	now := time.Now()
	for i := 0; i < 14; i++ {
		day := now.AddDate(0, 0, -i)
		start := day.Format("2006-01-02 00:00:00")
		end := day.Format("2006-01-02 23:59:59")
		
		cmd := exec.Command("git", "-C", repoPath, "rev-list", "--count", "--all", "--since=\""+start+"\"", "--until=\""+end+"\"")
		output, _ := cmd.CombinedOutput()
		var count int
		fmt.Sscanf(string(output), "%d", &count)
		history[13-i] = count
	}
	res, _ := json.Marshal(history)
	return string(res)
}

func getLastCommits(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "log", "--all", "-10", "--format=%H|%an|%cr|%s|%d")
	output, _ := cmd.CombinedOutput()
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	var results []string
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) < 5 { continue }
		
		refs := strings.TrimSpace(parts[4])
		if refs == "" || refs == "()" {
			hash := parts[0]
			// More reliable branch check in normal clones
			branchCmd := exec.Command("git", "-C", repoPath, "branch", "-a", "--contains", hash)
			branchOut, _ := branchCmd.CombinedOutput()
			bLines := strings.Split(strings.TrimSpace(string(branchOut)), "\n")
			
			if len(bLines) > 0 {
				for _, b := range bLines {
					b = strings.TrimSpace(strings.TrimPrefix(b, "*"))
					if !strings.Contains(b, "HEAD") && b != "" {
						branch := strings.TrimPrefix(b, "remotes/origin/")
						parts[4] = "(" + branch + ")"
						break
					}
				}
			}
		}
		results = append(results, strings.Join(parts, "|"))
	}
	return strings.Join(results, "\n")
}

func calculateHealthScore(meta GitHubMeta, historyStr string, lastCommitDate time.Time) int {
	score := 50
	var history []int
	json.Unmarshal([]byte(historyStr), &history)
	activeDays := 0
	for _, count := range history { if count > 0 { activeDays++ } }
	score += activeDays * 3
	daysSinceLast := int(time.Since(lastCommitDate).Hours() / 24)
	if daysSinceLast > 30 { score -= 10 }
	if daysSinceLast > 90 { score -= 20 }
	if daysSinceLast > 365 { score -= 30 }
	if meta.StargazersCount > 1000 { score += 5 }
	if meta.StargazersCount > 10000 { score += 5 }
	if score < 0 { score = 0 }
	if score > 100 { score = 100 }
	return score
}

func updateStatus(id int, status, errMsg string) {
	db.Exec("UPDATE repositories SET status = ?, error_message = ? WHERE id = ?", status, errMsg, id)
	broadcastStatus(id, status, errMsg)
}
