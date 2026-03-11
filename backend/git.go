package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func syncRepo(id int, url, name string) {
	updateStatus(id, "syncing", "")
	
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

	source, err := GetSource(url)
	var meta Metadata
	if err == nil {
		meta, _ = source.GetMetadata(url)
		if meta.Username != "" {
			downloadAvatar(meta.AvatarURL, meta.Username)
		}

		// Sync Wiki
		if wikiURL, exists := source.GetWikiURL(url); exists {
			syncWiki(wikiURL, name)
		}

		// Sync Non-Git Metadata (Issues, Releases)
		metadataPath := filepath.Join("./data", name, "metadata")
		source.SyncIssues(url, metadataPath)
		source.SyncReleases(url, metadataPath)
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
		time.Now(), lastCommits, meta.Stars, meta.Forks, meta.OpenIssues, history, score, defaultBranch, id)
	
	broadcastStatus(id, "synced", "")
}

func syncWiki(url string, name string) {
	wikiPath := filepath.Join("./data", name, "wiki")
	if _, err := os.Stat(wikiPath); os.IsNotExist(err) {
		// Try to clone, but don't fail if wiki doesn't exist (returns 128)
		cmd := exec.Command("git", "clone", url, wikiPath)
		cmd.Run()
	} else {
		exec.Command("git", "-C", wikiPath, "fetch", "--all", "--tags", "--force").Run()
	}
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

func calculateHealthScore(meta Metadata, historyStr string, lastCommitDate time.Time) int {
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
	if meta.Stars > 1000 { score += 5 }
	if meta.Stars > 10000 { score += 5 }
	if score < 0 { score = 0 }
	if score > 100 { score = 100 }
	return score
}

func updateStatus(id int, status, errMsg string) {
	db.Exec("UPDATE repositories SET status = ?, error_message = ? WHERE id = ?", status, errMsg, id)
	broadcastStatus(id, status, errMsg)
}
