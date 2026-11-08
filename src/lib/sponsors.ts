import type {ShapeName} from '@/components/ExpressiveShape/paths';

export interface Sponsor {
  name: string;
  login?: string;
  avatarUrl?: string;
  url?: string;
  source: 'github' | 'kofi';
}

export interface SponsorsFeed {
  current: Sponsor[];
  past: Sponsor[];
  updatedAt: string;
}

export const GITHUB_SPONSORS_URL = 'https://github.com/sponsors/AvengeMedia';
export const KOFI_URL = 'https://ko-fi.com/duxlabs';

// Ko-fi supporters before the webhook, from the dashboard export (Nov 2026). The export
// doesn't say who tipped privately, so they're counted, not named. Bump it now and then.
export const KOFI_SUPPORTER_COUNT = 115;

// Shapes that keep most of a square avatar visible. Assigned by index so SSR and client agree.
export const AVATAR_SHAPES: ShapeName[] = ['Cookie9Sided', 'Clover4Leaf', 'Cookie12Sided', 'Sunny', 'SoftBurst', 'Gem'];

export class SponsorsError extends Error {
  constructor(readonly warmingUp: boolean) {
    super(warmingUp ? 'sponsors cache is warming up' : 'failed to load sponsors');
  }
}

function apiOrigin() {
  const isDev = process.env.NODE_ENV === 'development';
  return isDev ? 'http://localhost:8337' : 'https://api.duxlabs.dev';
}

export async function fetchSponsors(signal?: AbortSignal): Promise<SponsorsFeed> {
  // The edge adds max-age=600 even to 503s, so skip the browser cache or "Try again" replays the error.
  const response = await fetch(`${apiOrigin()}/sponsors`, {signal, cache: 'no-cache'});
  if (!response.ok) throw new SponsorsError(response.status === 503);
  const data: Partial<SponsorsFeed> = await response.json();
  return {current: data.current ?? [], past: data.past ?? [], updatedAt: data.updatedAt ?? ''};
}
