import { apiFetch, toastHandler } from "$lib/api.svelte";
import { endpoint } from "./endpoints";
import type { Settings } from "./types";

export class SettingsAPI {
  async getSettings(): Promise<Settings> {
    try {
      const url = endpoint.GET_SETTINGS();
      const res = await apiFetch(url);
      if (res.ok) {
        const settings = await res.json();
        return settings;
      }
      return {
        github_token_set: false,
        gitlab_url: '',
        gitlab_token_set: false,
        gitea_url: '',
        gitea_token_set: false,
        export_destination: '',
      }
    } catch (e: any) {
      toastHandler.showToast('Failed to get settings', 'error');
      return {
        github_token_set: false,
        gitlab_url: '',
        gitlab_token_set: false,
        gitea_url: '',
        gitea_token_set: false,
        export_destination: '',
      }
    }
  }

  async setSettings(github: string, gitlabURL: string, gitlabToken: string, giteaURL: string, giteaToken: string, exportDest: string): Promise<boolean> {
    try {
      const url = endpoint.SET_SETTINGS();
      const res = await apiFetch(url, {
        method: 'PATCH',
        body: JSON.stringify({
          'github_token': github,
          'gitlab_url': gitlabURL,
          'gitlab_token': gitlabToken,
          'gitea_url': giteaURL,
          'gitea_token': giteaToken,
          'export_destination': exportDest,
        })
      });

      return res.ok;
    } catch (e: any) {
      return false;
    }
  }
}
