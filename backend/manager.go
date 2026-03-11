package main

import (
	"log"
	"sync"
)

type SyncTask struct {
	ID   int
	URL  string
	Name string
}

type SyncManager struct {
	tasks       chan SyncTask
	activeTasks map[int]bool
	mu          sync.Mutex
	workerCount int
}

var manager *SyncManager

func initSyncManager(workerCount int) {
	manager = &SyncManager{
		tasks:       make(chan SyncTask, 100),
		activeTasks: make(map[int]bool),
		workerCount: workerCount,
	}

	for i := 0; i < workerCount; i++ {
		go manager.worker(i)
	}
	log.Printf("[MANAGER] Started SyncManager with %d workers", workerCount)
}

func (m *SyncManager) Enqueue(id int, url, name string) {
	m.mu.Lock()
	if m.activeTasks[id] {
		m.mu.Unlock()
		log.Printf("[MANAGER] Task for repo %s already in queue or active, skipping duplicate.", name)
		return
	}
	m.activeTasks[id] = true
	m.mu.Unlock()

	m.tasks <- SyncTask{ID: id, URL: url, Name: name}
	log.Printf("[MANAGER] Enqueued sync for %s", name)
}

func (m *SyncManager) worker(id int) {
	for task := range m.tasks {
		log.Printf("[WORKER %d] Starting sync for %s", id, task.Name)
		syncRepo(task.ID, task.URL, task.Name)
		log.Printf("[WORKER %d] Finished sync for %s", id, task.Name)

		m.mu.Lock()
		delete(m.activeTasks, task.ID)
		m.mu.Unlock()
	}
}
