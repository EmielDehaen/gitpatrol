import { type Repository, type Incident, type User, type Toast } from './types';

export const API_URL = 'http://localhost:8080';

class GitPatrolAPI {
  isAuthenticated = $state(false);
  needsBootstrap = $state(false);
  user = $state<User | null>(null);
  repositories = $state<Repository[]>([]);
  incidents = $state<Incident[]>([]);
  authLoading = $state(true);
  toasts = $state<Toast[]>([]);
  private toastId = 0;

  constructor() {
    this.checkAuth();
  }

  showToast(message: string, type: 'success' | 'error' | 'info' = 'info') {
    const id = this.toastId++;
    this.toasts = [...this.toasts, { id, message, type }];
    setTimeout(() => {
      this.toasts = this.toasts.filter(t => t.id !== id);
    }, 5000);
  }

  async apiFetch(endpoint: string, options: RequestInit = {}) {
    const res = await fetch(`${API_URL}${endpoint}`, {
      ...options,
      credentials: 'include'
    });
    if (res.status === 401 && this.isAuthenticated) {
      this.handleLogout();
      this.showToast('Session expired. Please login again.', 'error');
      throw new Error('Unauthorized');
    }
    return res;
  }

  async checkAuth() {
    try {
      const res = await fetch(`${API_URL}/api/auth/status`, { credentials: 'include' });
      const data = await res.json();
      this.needsBootstrap = data.needs_bootstrap || false;
      this.isAuthenticated = data.authenticated || false;
      
      if (this.isAuthenticated) {
        const meRes = await this.apiFetch('/api/me');
        if (meRes.ok) {
          this.user = await meRes.json();
        }
        this.fetchRepos();
        this.fetchIncidents();
      }
    } catch (e) {
      console.error('Auth check failed', e);
    } finally {
      this.authLoading = false;
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
      this.showToast(this.needsBootstrap ? 'System bootstrapped!' : 'Welcome back.', 'success');
      await this.checkAuth();
      return true;
    } else {
      const data = await res.json();
      this.showToast(data.error || 'Authentication failed.', 'error');
      return false;
    }
  }

  async handleLogout() {
    await fetch(`${API_URL}/api/auth/logout`, { method: 'POST', credentials: 'include' });
    this.isAuthenticated = false;
    this.user = null;
    this.repositories = [];
    this.showToast('Logged out.', 'info');
  }

  async fetchRepos() {
    try {
      const res = await this.apiFetch('/api/repositories');
      if (res.ok) {
        this.repositories = await res.json();
      }
    } catch (e) {}
  }

  async fetchIncidents() {
    try {
      const res = await this.apiFetch('/api/incidents');
      if (res.ok) {
        this.incidents = await res.json();
      }
    } catch (e) {}
  }

  async clearIncidents() {
    try {
      const res = await this.apiFetch('/api/incidents', { method: 'DELETE' });
      if (res.ok) {
        this.incidents = [];
        this.showToast('All incidents cleared.', 'info');
      }
    } catch (e) {}
  }
}

export const api = new GitPatrolAPI();
