# REST API Reference 🔌

GitPatrol provides a RESTful API that can be used to integrate with third-party tools or write custom plugins.

## Authentication
GitPatrol uses **JWT (JSON Web Tokens)** stored in Http-Only cookies for session management.

- **Cookie Name:** `session_token`
- **Security:** In production, ensure `GP_SECURE_COOKIE=true` is set.

---

## 🌍 Public Endpoints

### Get Auth Status
Returns whether the system needs bootstrapping and if the current session is valid.
- **URL:** `/api/auth/status`
- **Method:** `GET`
- **Response:**
  ```json
  {
    "needs_bootstrap": false,
    "logged_in": true
  }
  ```

### Get Health Badge
Returns a dynamic SVG health badge for a specific repository. No authentication required.
- **URL:** `/api/repositories/:id/badge`
- **Method:** `GET`
- **Returns:** SVG Image

### Bootstrap / Register
Creates the first administrator account. Only available if `needs_bootstrap` is `true`.
- **URL:** `/api/auth/register`
- **Method:** `POST`
- **Body:** `{ "username": "...", "password": "..." }`

### Login
- **URL:** `/api/auth/login`
- **Method:** `POST`
- **Body:** `{ "username": "...", "password": "..." }`

---

## 🔐 Protected Endpoints
*These endpoints require a valid `session_token` cookie.*

### Get Current User
- **URL:** `/api/me`
- **Method:** `GET`

### Update User / Password
- **URL:** `/api/user`
- **Method:** `PATCH`
- **Body:** `{ "username": "...", "old_password": "...", "new_password": "..." }`

### List Repositories
- **URL:** `/api/repositories`
- **Method:** `GET`

### Add Repository
- **URL:** `/api/repositories`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "name": "My Repo",
    "url": "https://github.com/user/repo",
    "interval_minutes": 60,
    "auto_patrol": 1
  }
  ```

### Update Repository Configuration
- **URL:** `/api/repositories/:id`
- **Method:** `PATCH`
- **Body:** `{ "interval_minutes": 120, "auto_patrol": 0 }`

### Sync Now
Triggers an immediate synchronization task.
- **URL:** `/api/repositories/:id/sync`
- **Method:** `POST`

### Get README
Returns the README content rendered as Markdown (or raw).
- **URL:** `/api/repositories/:id/readme`
- **Method:** `GET`

### Get Incidents
Returns a list of recent synchronization failures.
- **URL:** `/api/incidents`
- **Method:** `GET`

### Clear Incidents
Marks all incidents as resolved.
- **URL:** `/api/incidents`
- **Method:** `DELETE`

---

## 🛠️ WebSocket
Real-time updates are streamed via WebSockets.
- **URL:** `/ws`
- **Event:** Sends a generic message whenever a repository status changes or a new incident is created.
