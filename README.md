# GitPatrol ⚡

A minimalist, high-performance Git repository mirroring and health-monitoring dashboard.

## Tech Stack
- **Backend:** Go (Golang) + SQLite
- **Frontend:** SvelteKit + Bun
- **Deployment:** Docker / Docker Compose

## Features
- **Automatic Mirroring:** Keeps a full `--mirror` clone of your favorite repositories.
- **Scheduled Fetching:** Configurable sync intervals (per-repo).
- **Health Monitoring:** Detects if remote repositories still exist (404 check) or have become stagnant.
- **Real-time Updates:** WebSocket-driven dashboard for live sync status and git logs.
- **Efinity Vibe:** Minimalist, fast, and visually polished dashboard.

## How to Run

### Using Docker (Recommended)
1. Clone this repository.
2. Run `docker-compose up -d`.
3. Open `http://localhost:3000` in your browser.

### Development Mode

**Backend:**
```bash
cd backend
go run .
```

**Frontend:**
```bash
cd frontend
bun install
bun run dev
```

## Configuration
All configuration is handled via the dashboard UI or the `.env` file in the frontend.
