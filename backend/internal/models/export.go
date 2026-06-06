package models

type ExportRequest struct {
	RepoID      int    `json:"repo_id"`
	Destination string `json:"destination"` // e.g., "gitea"
}

type ExportResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	DestinationURL string `json:"destination_url,omitempty"`
}

type DestinationConfig struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}
