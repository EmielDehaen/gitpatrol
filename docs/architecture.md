# Architecture ⚡

GitPatrol follows a modern decoupled architecture designed for high performance and low-overhead repository mirroring.

## System Overview
The system consists of three main components:

1.  **Backend (Go):** A high-concurrency engine built with Go, responsible for:
    - Managing the SQLite database.
    - Interfacing with the Git CLI (or `go-git` in the future) for mirroring.
    - Exposing a REST API for the frontend.
    - Streaming real-time updates via WebSockets.
2.  **Frontend (SvelteKit):** A reactive dashboard powered by SvelteKit and Bun, providing:
    - A visual overview of all mirrored repositories.
    - Health score visualizations.
    - Direct access to repository metadata (READMEs, logs).
    - Real-time status indicators.
3.  **Data Persistence:**
    - **SQLite:** Stores configuration, health scores, and metadata.
    - **Local Disk:** Stores the physical Git clones and mirrored objects.

## Data Flow
1.  **Repository Added:** User submits a URL via the UI.
2.  **Initial Clone:** The backend performs a full clone of the repository.
3.  **Metadata Extraction:** The backend pulls the latest stars, forks, and issues from the provider's API (e.g., GitHub).
4.  **Scheduled Sync:** Based on the user-defined interval, the backend fetches updates from the remote to keep the mirror up to date.
5.  **Health Score:** On each sync, the system recalculates the "Health Score" based on activity, stars, and remote availability.

## Design Principles
- **Minimalist:** No heavy frameworks where not needed.
- **Offline-First:** All repo data is mirrored locally; the dashboard works without internet for cloned assets.
- **Resilient:** Uses SQLite for zero-config database management.
