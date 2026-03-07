package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func syncRepo(id int, url, name string) {
	updateStatus(id, "syncing", "")
	
	repoPath := filepath.Join("./data", name)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		// Clone as mirror
		cmd := exec.Command("git", "clone", "--mirror", url, repoPath)
		if err := cmd.Run(); err != nil {
			updateStatus(id, "error", err.Error())
			return
		}
	} else {
		// Fetch updates
		cmd := exec.Command("git", "--git-dir="+repoPath, "fetch", "-p", "origin")
		if err := cmd.Run(); err != nil {
			updateStatus(id, "error", err.Error())
			return
		}
	}

	// Get last commit
	cmd := exec.Command("git", "--git-dir="+repoPath, "log", "-1", "--format=%H %cd")
	output, _ := cmd.CombinedOutput()
	
	updateStatusSuccess(id, string(output))
}

func updateStatus(id int, status, errMsg string) {
	db.Exec("UPDATE repositories SET status = ?, error_message = ? WHERE id = ?", status, errMsg, id)
	broadcastStatus(id, status, errMsg)
}

func updateStatusSuccess(id int, lastCommit string) {
	db.Exec("UPDATE repositories SET status = 'synced', last_sync = ?, last_commit = ?, error_message = '' WHERE id = ?", 
		time.Now(), lastCommit, id)
	broadcastStatus(id, "synced", "")
}
