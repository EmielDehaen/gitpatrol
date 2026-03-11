# GitPatrol ⚡

**GitPatrol** is a professional-grade, high-performance dashboard for Git repository mirroring and health monitoring. It serves as your "Code Insurance," ensuring that critical open-source assets remain available and active even if the remote hoster fails or the repository is deleted.

Built for the **Efinity** standard: minimalist, fast, and visually polished.

---

## ✨ Key Features
- **Professional Mirroring:** Maintains a full, accumulative archive of repositories.
- **Health Monitoring:** Intelligently calculates a "Health Score" based on activity, stars, and remote availability.
- **Multi-Provider Ready:** Built with a modular source system (GitHub, GitLab, and more).
- **Offline Discovery:** Explore READMEs and commit history directly from the mirrored local data.
- **Real-time Engine:** WebSocket-driven updates for live synchronization monitoring.

---

## 🛠️ Tech Stack
- **Backend:** Go (Golang 1.24) + SQLite
- **Frontend:** SvelteKit + Bun + Tailwind CSS 4.0
- **Deployment:** Docker & Docker Compose

---

## 📖 Documentation
Detailed documentation is available in the `docs/` folder:
- [**Architecture**](docs/architecture.md) - How the system works under the hood.
- [**Setup & Installation**](docs/setup.md) - How to get GitPatrol running.
- [**Roadmap**](docs/roadmap.md) - The vision and future features.

---

## 🚀 Quick Start (Docker)
1.  **Clone:** `git clone https://github.com/EmielDehaen/gitpatrol.git`
2.  **Run:** `docker-compose up -d`
3.  **Enjoy:** Open `http://localhost:3000`

---

## 🛡️ License
Licensed under the **PolyForm Non-Commercial License 1.0.0**. See [LICENSE](LICENSE) for details.

---
*Created with ⚡ by [Efinity](https://efinity.be)*
