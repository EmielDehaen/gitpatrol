import { SettingsAPI } from './settings/settingsApi';
import { type Repository, type Incident, type User, type Toast, type HealthStatus } from './types';

// Use hardcoded URL for dev, or detect from window for production/docker
export const API_URL = (typeof window !== 'undefined' && window.location.origin.includes(':3000'))
  ? 'http://localhost:8080'
  : 'http://localhost:8080'; // Default fallback

// Note: In a real production build, we might want to use window.location.origin

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
    // Deduplicate: don't show the same message+type if it's already visible
    if (this.toasts.some(t => t.message === message && t.type === type)) return;
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
  healthStatus = $state<HealthStatus | null>(null);
  authLoading = $state(true);
  settings = new SettingsAPI();
  private healthTimer: ReturnType<typeof setInterval> | null = null;



  setupHealthPolling() {
    if (this.healthTimer) clearInterval(this.healthTimer);
    this.healthTimer = setInterval(() => {
      if (this.isAuthenticated) this.fetchHealth();
    }, 60000); // Fallback polling every 1 minute
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
        this.setupHealthPolling();
      }
    } catch (e) {
      console.error('Auth check failed', e);
      this.healthStatus = null;
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
        this.healthStatus = null;
      }
    } catch (e) {
      this.healthStatus = null;
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
    } catch (e) {
      toastHandler.showToast('Network error: could not load repositories.', 'error');
    }
  }

  async fetchIncidents() {
    try {
      const res = await apiFetch('/api/incidents');
      if (res.ok) {
        this.incidents = await res.json();
      }
    } catch (e) {
      toastHandler.showToast('Network error: could not load incidents.', 'error');
    }
  }

  async clearIncidents() {
    try {
      const res = await apiFetch('/api/incidents', { method: 'DELETE' });
      if (res.ok) {
        this.incidents = [];
        toastHandler.showToast('All incidents cleared.', 'info');
      }
    } catch (e) {
      toastHandler.showToast('Network error: could not clear incidents.', 'error');
    }
  }
}

export const api = new GitPatrolAPI();
export const toastHandler = new ToastHandler();