export function getNormalizedUrl(url: string) {
  if (!url || url.length < 3) return '';
  let u = url.trim().replace(/\.git$/, '');
  if (u.startsWith('git@')) {
    u = u.replace(':', '/').replace('git@', 'https://');
  }
  if (u.includes('/tree/')) u = u.split('/tree/')[0];
  if (u.includes('/blob/')) u = u.split('/blob/')[0];
  if (!u.startsWith('http') && u.includes('/')) {
    const parts = u.split('/');
    if (parts.length === 2) u = 'https://github.com/' + u;
  }
  return u;
}

export function suggestName(url: string) {
  if (!url) return '';
  let cleanUrl = getNormalizedUrl(url);
  const parts = cleanUrl.split('/');
  let name = parts[parts.length - 1];
  return name ? name.charAt(0).toUpperCase() + name.slice(1) : '';
}

export function minutesToHuman(minutes: number): string {
  if (minutes <= 0) return '0m';
  const d = Math.floor(minutes / 1440);
  const h = Math.floor((minutes % 1440) / 60);
  const m = minutes % 60;
  
  let parts = [];
  if (d > 0) parts.push(`${d}d`);
  if (h > 0) parts.push(`${h}h`);
  if (m > 0) parts.push(`${m}m`);
  return parts.join(' ') || '0m';
}

export function humanToMinutes(str: string): number {
  const regex = /(?:(\d+)d)?\s*(?:(\d+)h)?\s*(?:(\d+)m)?/i;
  const match = str.match(regex);
  if (!match) return 60;
  
  const d = parseInt(match[1] || '0');
  const h = parseInt(match[2] || '0');
  const m = parseInt(match[3] || '0');
  
  const total = (d * 1440) + (h * 60) + m;
  return total > 0 ? total : 60;
}

export function getAvatarUrl(url: string, apiUrl: string) {
  const parts = url.replace('https://github.com/', '').split('/');
  if (parts.length > 0) return `${apiUrl}/avatars/${parts[0]}.png`;
  return '';
}

export function isSyncing(repo: any) {
  if (repo.status === 'syncing') return true;
  if (repo.auto_patrol === 0 || !repo.last_sync) return false;
  const lastSync = new Date(repo.last_sync).getTime();
  const nextSync = lastSync + repo.interval_minutes * 60000;
  return (nextSync - Date.now()) <= 0;
}

export function getProgress(repo: any) {
  if (!repo.last_sync || repo.status === 'syncing' || repo.auto_patrol === 0) return 0;
  const lastSync = new Date(repo.last_sync).getTime();
  const nextSync = lastSync + repo.interval_minutes * 60000;
  const total = repo.interval_minutes * 60000;
  const remaining = nextSync - Date.now();
  if (remaining <= 0) return 0;
  return 100 * (remaining / total);
}

export function getRemainingTime(repo: any) {
  if (repo.auto_patrol === 0) return 'Manual Patrol Only';
  if (!repo.last_sync || repo.status === 'syncing') return 'Syncing...';
  
  const lastSync = new Date(repo.last_sync).getTime();
  const nextSync = lastSync + repo.interval_minutes * 60000;
  const remainingMs = nextSync - Date.now();
  
  if (remainingMs <= 0) return 'Syncing...';
  
  const d = Math.floor(remainingMs / (1000 * 60 * 60 * 24));
  const h = Math.floor((remainingMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
  const m = Math.floor((remainingMs % (1000 * 60 * 60)) / (1000 * 60));
  const s = Math.floor((remainingMs % (1000 * 60)) / 1000);
  
  let parts = [];
  if (d > 0) parts.push(`${d}d`);
  if (h > 0) parts.push(`${h}h`);
  if (m > 0) parts.push(`${m}m`);
  if (s > 0 || parts.length === 0) parts.push(`${s}s`);
  
  return `Next sync in: ${parts.join(' ')}`;
}
