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
        gitlab_token_set: false,
      }
    } catch (e: any) {
      toastHandler.showToast('Failed to get settings', 'error');
      console.log(e);
      return {
        github_token_set: false,
        gitlab_token_set: false,
      }
    }
  }

  async setSettings(github: string, gitlab: string): Promise<boolean> {
    try {
      const url = endpoint.SET_SETTINGS();
      const res = await apiFetch(url, {
        method: 'PATCH',
        body: JSON.stringify({
          'github_token': github,
          'gitlab_token': gitlab,
        })
      });

      if (res.ok) {
        return true;
      }

      return false;
    } catch (e: any) {
      return false;
    }
  }
}