import { SettingsAPI } from './settings/settingsApi';
import { type Repository, type Incident, type User, type Toast } from './types';

export const API_URL = 'http://localhost:8080';

export async function apiFetch(endpoint: string, options: RequestInit = {}) {
  const res = await fetch(`${API_URL}${endpoint}`, {
    ...options,
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' }
  });
  if (res.status === 401 && api.isAuthenticated) {
    api.handleLogout();
    toastHandler.showToast('Session expired. Please login again.', 'error');
    throw new Error('Unauthorized');
  }
  return res;
}

class ToastHandler {
  toasts = $state<Toast[]>([]);
  private toastId = 0

  showToast(message: string, type: 'success' | 'error' | 'info' = 'info') {
    const id = this.toastId++;
    this.toasts = [...this.toasts, { id, message, type }];
    setTimeout(() => {
      this.toasts = this.toasts.filter(t => t.id !== id);
    }, 5000);
  }
}

class GitPatrolAPI {
  isAuthenticated = $state(false);
  needsBootstrap = $state(false);
  user = $state<User | null>(null);
  repositories = $state<Repository[]>([]);
  incidents = $state<Incident[]>([]);
  healthStatus = $state<{ status: string, checks: any } | null>(null);
  authLoading = $state(true);
  settings = new SettingsAPI();

  constructor() {
    this.checkAuth();
    setInterval(() => {
      if (this.isAuthenticated) this.fetchHealth();
    }, 30000);
  }

  async checkAuth() {
    try {
      const res = await fetch(`${API_URL}/api/auth/status`, { credentials: 'include' });
      const data = await res.json();
      this.needsBootstrap = data.needs_bootstrap || false;
      this.isAuthenticated = data.authenticated || false;

      if (this.isAuthenticated) {
        const meRes = await apiFetch('/api/me');
        if (meRes.ok) {
          this.user = await meRes.json();
        }
        this.fetchRepos();
        this.fetchIncidents();
        this.fetchHealth();
      }
    } catch (e) {
      console.error('Auth check failed', e);
      this.healthStatus = { status: 'offline', checks: {} };
    } finally {
      this.authLoading = false;
    }
  }

  async fetchHealth() {
    try {
      const res = await apiFetch('/api/health');
      if (res.ok) {
        this.healthStatus = await res.json();
      } else {
        this.healthStatus = { status: 'offline', checks: {} };
      }
    } catch (e) {
      this.healthStatus = { status: 'offline', checks: {} };
    }
  }

  async handleAuth(username: string, password: string) {
    const endpoint = this.needsBootstrap ? '/api/auth/register' : '/api/auth/login';
    const res = await fetch(`${API_URL}${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
      credentials: 'include'
    });

    if (res.ok) {
      toastHandler.showToast(this.needsBootstrap ? 'System bootstrapped!' : 'Welcome back.', 'success');
      await this.checkAuth();
      return true;
    } else {
      const data = await res.json();
      toastHandler.showToast(data.error || 'Authentication failed.', 'error');
      return false;
    }
  }

  async handleLogout() {
    await fetch(`${API_URL}/api/auth/logout`, { method: 'POST', credentials: 'include' });
    this.isAuthenticated = false;
    this.user = null;
    this.repositories = [];
    toastHandler.showToast('Logged out.', 'info');
  }

  async fetchRepos() {
    try {
      const res = await apiFetch('/api/repositories');
      if (res.ok) {
        this.repositories = await res.json();
      }
    } catch (e) { }
  }

  async fetchIncidents() {
    try {
      const res = await apiFetch('/api/incidents');
      if (res.ok) {
        this.incidents = await res.json();
      }
    } catch (e) { }
  }

  async clearIncidents() {
    try {
      const res = await apiFetch('/api/incidents', { method: 'DELETE' });
      if (res.ok) {
        this.incidents = [];
        toastHandler.showToast('All incidents cleared.', 'info');
      }
    } catch (e) { }
  }
}

export const api = new GitPatrolAPI();
export const toastHandler = new ToastHandler();