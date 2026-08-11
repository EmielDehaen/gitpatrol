export const API_URL = import.meta.env.VITE_API_URL || '';

export async function apiFetch(endpoint: string, options: RequestInit = {}) {
  const res = await fetch(`${API_URL}${endpoint}`, {
    ...options,
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' }
  });
  if (res.status === 401) {
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('gitpatrol:unauthorized'));
    }
    throw new Error('Unauthorized');
  }
  return res;
}
