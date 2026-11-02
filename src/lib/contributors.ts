export interface CodeContributor {
  login: string;
  avatarUrl?: string;
  url: string;
  contributions: number;
  repos: string[];
}

export interface TranslatorLanguage {
  projectId: string;
  project: string;
  code: string;
  name: string;
}

export interface Translator {
  name: string;
  proofreader?: boolean;
  languages: TranslatorLanguage[];
}

export interface ContributorsFeed {
  code: CodeContributor[];
  translators: Translator[];
  updatedAt: string;
}

export interface LanguageCredit {
  code: string;
  name: string;
  translators: string[];
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
  return {code: data.code ?? [], translators: data.translators ?? [], updatedAt: data.updatedAt ?? ''};
}

// One credit per language across every POEditor project, most translators first.
export function languageCredits(translators: Translator[]): LanguageCredit[] {
  const byCode = new Map<string, {name: string; people: Set<string>}>();
  for (const t of translators) {
    for (const lang of t.languages) {
      const entry = byCode.get(lang.code) ?? {name: lang.name, people: new Set<string>()};
      entry.people.add(t.name);
      byCode.set(lang.code, entry);
    }
  }
  return [...byCode.entries()]
    .map(([code, {name, people}]) => ({
      code,
      name,
      translators: [...people].sort((a, b) => a.localeCompare(b)),
    }))
    .sort((a, b) => b.translators.length - a.translators.length || a.name.localeCompare(b.name));
}

// GitHub serves full-size avatars by default; ask for roughly 2x the rendered size.
export function sizedAvatar(url: string | undefined, px: number): string | undefined {
  if (!url) return undefined;
  return `${url}${url.includes('?') ? '&' : '?'}s=${px * 2}`;
}
