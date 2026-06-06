export interface Repository {
  id: number;
  name: string;
  url: string;
  interval_minutes: number;
  last_sync: string;
  status: string;
  last_commit: string;
  error_message: string;
  stars: number;
  forks: number;
  open_issues: number;
  commit_history: string;
  health_score: number;
  default_branch: string;
  auto_patrol: number;
  progress?: number;
}

export interface Incident {
  id: number;
  repo_id: number | null;
  repo_name: string;
  message: string;
  created_at: string;
  resolved: number;
}

export interface User {
  id: number;
  username: string;
}

export interface Toast {
  id: number;
  message: string;
  type: 'success' | 'error' | 'info';
}

export interface Commit {
  hash: string;
  author: string;
  date: string;
  message: string;
  branch: string;
}
