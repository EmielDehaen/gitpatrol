# Setup & Installation ⚙️

GitPatrol is designed for simple, robust deployment within home labs or corporate infrastructure.

## Prerequisites
- **Docker & Docker Compose** (Recommended)
- **Git CLI** (1.7+ or higher)
- **Go 1.24+** (for manual backend builds)
- **Bun** (for manual frontend development)

---

## 🐋 Docker Installation (Recommended)
Running GitPatrol via Docker Compose is the easiest way to ensure all dependencies are correctly configured.

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/EmielDehaen/gitpatrol.git
    cd gitpatrol
    ```
2.  **Environment Configuration:**
    Create a `.env` file in the root directory (refer to the `.env.example` in each subfolder if needed).
3.  **Spin up the containers:**
    ```bash
    docker-compose up -d
    ```
4.  **Access the Dashboard:**
    Open `http://localhost:3000` to start mirroring your repositories.

---

## 🛠️ Manual Development Setup

### Backend (Go)
1.  **Enter the directory:**
    ```bash
    cd backend
    ```
2.  **Run the application:**
    ```bash
    go run .
    ```
    The backend runs on port `:8080` by default.

### Frontend (SvelteKit)
1.  **Enter the directory:**
    ```bash
    cd frontend
    ```
2.  **Install dependencies:**
    ```bash
    bun install
    ```
3.  **Start the development server:**
    ```bash
    bun run dev
    ```
    The dashboard will be available at `http://localhost:5173`.

## 📦 Data Volumes
If running in Docker, the following folders should be persisted:
- `/root/gitpatrol/backend/db/`: SQLite database.
- `/root/gitpatrol/backend/data/`: Git mirrored repositories and avatars.
