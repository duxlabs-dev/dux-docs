export interface CodeContributor {
  login: string;
  avatarUrl?: string;
  url: string;
  contributions: number;
  repos: string[];
}

// Translators are counted, never named: they have not agreed to be listed.
export interface TranslationLanguage {
  code: string;
  name: string;
  translators: number;
}

export interface ContributorsFeed {
  code: CodeContributor[];
  languages: TranslationLanguage[];
  translatorCount: number;
  updatedAt: string;
}

export const POEDITOR_JOIN_URL = 'https://poeditor.com/join/project/kyeCPTpmJM';
export const GITHUB_ORG_URL = 'https://github.com/duxlabs-dev';

export class ContributorsError extends Error {
  constructor(readonly warmingUp: boolean) {
    super(warmingUp ? 'contributors cache is warming up' : 'failed to load contributors');
  }
}

function apiOrigin() {
  const isDev = process.env.NODE_ENV === 'development';
  return isDev ? 'http://localhost:8337' : 'https://api.duxlabs.dev';
}

export async function fetchContributors(signal?: AbortSignal): Promise<ContributorsFeed> {
  // The edge adds max-age=600 even to 503s, so skip the browser cache or "Try again" replays the error.
  const response = await fetch(`${apiOrigin()}/contributors`, {signal, cache: 'no-cache'});
  if (!response.ok) throw new ContributorsError(response.status === 503);
  const data: Partial<ContributorsFeed> = await response.json();
  return {
    code: data.code ?? [],
    languages: data.languages ?? [],
    translatorCount: data.translatorCount ?? 0,
    updatedAt: data.updatedAt ?? '',
  };
}

// GitHub serves full-size avatars by default; ask for roughly 2x the rendered size.
export function sizedAvatar(url: string | undefined, px: number): string | undefined {
  if (!url) return undefined;
  return `${url}${url.includes('?') ? '&' : '?'}s=${px * 2}`;
}
