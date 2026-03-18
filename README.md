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

## 🐳 Docker Deployment
GitPatrol offers two ways to run with Docker:

### 1. Production (Recommended)
Uses the optimized **single-binary image** (approx. 30MB). This is the fastest and most resource-efficient way to run GitPatrol.

```bash
docker-compose -f docker-compose.production.yml up -d
```
*The image `efinityhub/gitpatrol:latest` is automatically updated on every release.*

### 2. Development
If you want to modify the source code and see changes in real-time, use the standard compose file:

```bash
docker-compose up -d
```

---

## 🚀 Professional Installation (Linux / VPS / Laptop)
The fastest way to deploy GitPatrol on any Debian-based system. This script automatically detects your architecture (AMD64/ARM64) and sets up a **systemd service**.

```bash
curl -fsSL https://raw.githubusercontent.com/efinityhub/gitpatrol/main/scripts/install.sh | sudo bash
```

---

## ⚡ Proxmox LXC Deployment
GitPatrol is fully compatible with the **Community-Scripts (tteck)** standard. Use the dedicated provisioner to create a fresh, optimized LXC container in seconds.

```bash
bash -c "$(curl -fsSL https://raw.githubusercontent.com/efinityhub/gitpatrol/main/proxmox/ct/gitpatrol.sh)"
```
*This script will handle container creation, resource allocation (RAM/CPU), and the full automated installation.*

---

## 🐳 Quick Start (Docker)
If you prefer container isolation via Docker, use the provided compose file:

1.  **Clone:** `git clone https://github.com/efinityhub/gitpatrol.git`
2.  **Run:** `docker-compose up -d`
3.  **Enjoy:** Open `http://localhost:3000` and follow the bootstrap instructions.

---

## 🛠️ Tech Stack & Single Binary
- **Backend:** Go (Golang 1.24) + Echo + SQLite
- **Frontend:** SvelteKit 5 + Tailwind CSS 4.0
- **Industrial Design:** The production build is a **single, self-contained binary** (approx. 17MB) with the entire frontend embedded. No Node.js or external dependencies are required to run, making it extremely lightweight for homelab environments.

---

## 🛡️ License
Licensed under the **PolyForm Non-Commercial License 1.0.0**. See [LICENSE](LICENSE) for details.

---
*Created with ⚡ by [Efinity](https://efinity.be)*
