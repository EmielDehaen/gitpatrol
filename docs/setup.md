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
    The backend automatically generates secrets on the first run, but you can configure the following in your environment or a `.env` file:
    | Variable | Description | Default |
    |----------|-------------|---------|
    | `JWT_SECRET` | Secret key for JWT tokens. | Randomly generated |
    | `PASSWORD_PEPPER` | Pepper used for password hashing. | Randomly generated |
    | `GP_SECURE_COOKIE` | Set to `true` to enable the Secure flag on cookies. | `false` |

3.  **Spin up the containers:**
    ```bash
    docker-compose up -d
    ```
4.  **Bootstrap the System:**
    Open `http://localhost:3000`. If it's your first time, you will be prompted to create the primary administrator account.

---

## 🛠️ Manual Development Setup

### Backend (Go)
1.  **Enter the directory:** `cd backend`
2.  **Run the application:** `go run .`
    The backend runs on port `:8080` by default and stores data in `./db` and `./data`.

### Frontend (SvelteKit)
1.  **Enter the directory:** `cd frontend`
2.  **Install dependencies:** `bun install`
3.  **Start the development server:** `bun run dev`
    The dashboard will be available at `http://localhost:5173`. Ensure it points to the correct `PUBLIC_API_URL`.

## 📦 Data Volumes
If running in Docker, ensure the following volumes are mapped to persistent storage:
- `/root/db/`: Contains `gitpatrol.db` (SQLite) and `gitpatrol.env`.
- `/root/data/`: Contains physical Git mirrors, avatars, and archived metadata.
