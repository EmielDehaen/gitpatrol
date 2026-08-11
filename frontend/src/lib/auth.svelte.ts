import { API_URL, apiFetch } from './client';
import { toastHandler } from './toast.svelte';
import type { User } from './types';

class AuthStore {
  isAuthenticated = $state(false);
  needsBootstrap = $state(false);
  user = $state<User | null>(null);
  authLoading = $state(true);

  constructor() {
    if (typeof window !== 'undefined') {
      window.addEventListener('gitpatrol:unauthorized', () => {
        if (this.isAuthenticated) {
          this.handleLogout();
          toastHandler.showToast('Session expired. Please login again.', 'error');
        }
      });
    }
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
    window.dispatchEvent(new CustomEvent('gitpatrol:logged_out'));
    toastHandler.showToast('Logged out.', 'info');
  }
}

export const authStore = new AuthStore();
