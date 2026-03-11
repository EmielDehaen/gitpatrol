# Core Features 🚀

GitPatrol provides a suite of tools for repository mirroring and monitoring, designed for high-concurrency and reliability.

## 📦 Professional Mirroring
- **Accumulative Fetching:** Uses `git fetch --all --tags --force` to ensure all remote objects are pulled locally.
- **Auto-Mirroring:** Periodically polls repositories for changes based on configurable intervals.
- **Avatar Archiving:** Automatically downloads and mirrors provider avatars (e.g., GitHub profile pictures) for offline display.

## 📊 Health Monitoring
The "Health Score" is a unique metric calculated on each sync to help you assess the vitality of a project:
- **Activity Check:** Analyzes commit frequency over the last 14 days.
- **Community Stats:** Tracks Stars, Forks, and Open Issues via provider APIs.
- **Staleness Analysis:** Decreases score if the last commit was more than 30, 90, or 365 days ago.
- **Remote Check:** Verifies if the repository still exists at its remote URL.

## 🖥️ Modern Dashboard
- **Real-Time Updates:** Uses WebSockets to stream sync progress and logs directly to the browser.
- **Offline README Preview:** View a repository's README directly from the mirrored disk without connecting to the remote.
- **Commit History Visualization:** See a 14-day activity sparkline for every repository at a glance.
- **Status Indicators:** Clear visual cues for `pending`, `syncing`, `synced`, or `error` states.

## 🔌 Multi-Provider Support
- **GitHub:** Deep integration for metadata fetching.
- **GitLab:** Core support for repository mirroring.
- **Generic Git:** Supports any valid Git URL for mirroring (metadata coming soon).
