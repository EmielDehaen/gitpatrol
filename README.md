# GitPatrol ⚡

**GitPatrol** is a professional-grade, high-performance dashboard for Git repository mirroring and health monitoring. It serves as your "Code Insurance," ensuring that critical open-source assets remain available and active even if the remote hoster fails or the repository is deleted.

Built for the **Efinity** standard: minimalist, fast, and visually polished.

---

## ✨ Key Features
- **Professional Mirroring:** Maintains a full, accumulative archive of repositories using `--mirror` clones.
- **Intel Extraction:** Locally archives Issues, Release Notes, and Wikis for offline access.
- **Health Monitoring:** Intelligently calculates a "Health Score" based on activity, stars, and remote availability.
- **Dynamic Badges:** Generate SVG health badges for your mirrored assets.
- **Multi-Provider Ready:** Built with a modular source system (GitHub, GitLab, and generic Git).
- **Asset Service:** Serves README images and documentation directly from the local Git object store.
- **Real-time Engine:** WebSocket-driven updates for live synchronization monitoring.

---

## 🔐 Community Edition Security
- **Single-User Admin:** Simple, secure administrative access for your homelab or team.
- **Bootstrap Ready:** Just run the container and create your admin account on the first visit.
- **Secure Sessions:** JWT-based authentication using Http-Only cookies with configurable security flags.

---

## 🛠️ Tech Stack
- **Backend:** Go (Golang 1.24) + Echo + SQLite
- **Frontend:** SvelteKit + Bun + Tailwind CSS 4.0
- **Deployment:** Docker & Docker Compose

---

## 📖 Documentation
Detailed documentation is available in the `docs/` folder:
- [**Introduction**](docs/introduction.md) - What is GitPatrol and why use it.
- [**Features**](docs/features.md) - Deep dive into core functionality.
- [**API Reference**](docs/api.md) - Integration guide for plugins and 3rd party tools.
- [**Architecture**](docs/architecture.md) - How the system works under the hood.
- [**Setup & Installation**](docs/setup.md) - How to get GitPatrol running.

---

## 🚀 Quick Start (Docker)
1.  **Clone:** `git clone https://github.com/EmielDehaen/gitpatrol.git`
2.  **Run:** `docker-compose up -d`
3.  **Enjoy:** Open `http://localhost:3000` and follow the bootstrap instructions.

---

## 🛡️ License
Licensed under the **PolyForm Non-Commercial License 1.0.0**. See [LICENSE](LICENSE) for details.

---
*Created with ⚡ by [Efinity](https://efinity.be)*
