import React, { useState, useEffect } from 'react';
import Layout from '@theme/Layout';
import Head from '@docusaurus/Head';
import { useLocation, useHistory } from '@docusaurus/router';
import {ArrowDownward, ArrowUpward, Download, Forum, GridView, KeyboardArrowDown, ThumbUp, ViewList} from 'material-symbols/rounded';
import LandingBackground from '../components/LandingBackground';
import {IconButton} from '@/components/ui/icon-button';
import {Segmented} from '@/components/ui/segmented';
import styles from './plugins.module.css';

interface Plugin {
  id: string;
  name: string;
  capabilities: string[];
  category: string;
  repo: string;
  author: string;
  firstParty: boolean;
  description: string;
  dependencies: string[];
  compositors: string[];
  distro: string[];
  screenshot?: string;
  previewUrl?: string;
  version?: string;
  icon?: string;
  permissions?: string[];
  requires_dms?: string;
  updated_at: string;
  upvotes: number;
  issueUrl?: string;
  status?: string[];
  similar?: string[];
}

interface ThemeVariantOption {
  id: string;
  name: string;
  dark?: Record<string, string>;
  light?: Record<string, string>;
}

interface ThemeFlavor {
  id: string;
  name: string;
  dark?: Record<string, string>;
  light?: Record<string, string>;
}

interface ThemeAccent {
  id: string;
  name: string;
  [flavorId: string]: string | Record<string, string>;
}

interface ThemeModeDefaults {
  flavor: string;
  accent: string;
}

interface ThemeVariants {
  type?: 'multi';
  default?: string;
  options?: ThemeVariantOption[];
  defaults?: { dark?: ThemeModeDefaults; light?: ThemeModeDefaults };
  flavors?: ThemeFlavor[];
  accents?: ThemeAccent[];
}

interface ThemeWCAGGroup {
  level: string;
  minRatio: number;
  worstPair?: string[];
}

interface ThemeWCAGBreakdown {
  name: string;
  mode: string;
  level: string;
  bodyLevel: string;
}

interface ThemeWCAGMode {
  level: string;
  minRatio: number;
  worstPair?: string[];
  body?: ThemeWCAGGroup;
  accent?: ThemeWCAGGroup;
  nonText?: ThemeWCAGGroup;
  variants?: Record<string, string>;
  breakdown?: ThemeWCAGBreakdown[];
}

interface ThemeWCAG {
  level: string;
  dark?: ThemeWCAGMode;
  light?: ThemeWCAGMode;
}

interface Theme {
  id: string;
  name: string;
  author: string;
  description: string;
  version: string;
  dark: Record<string, string>;
  light: Record<string, string>;
  variants?: ThemeVariants;
  wcag?: ThemeWCAG;
  previewUrl: string;
  updated_at: string;
}

interface PluginsResponse {
  plugins: Plugin[];
  count: number;
}

interface ThemesResponse {
  themes: Theme[];
  count: number;
}

type ActiveTab = 'plugins' | 'themes';

const pluginTypes = [
  { id: 'all', label: 'All Types' },
  { id: 'dankbar-widget', label: 'Bar Widget' },
  { id: 'control-center', label: 'Control Center' },
  { id: 'launcher', label: 'Launcher' },
  { id: 'daemon', label: 'Daemon' },
  { id: 'desktop-widget', label: 'Desktop Widget' },
  { id: 'dash-tab', label: 'Dash Tab' },
  { id: 'dash-card', label: 'Dash Card' },
];

const pluginTypeLabel = (capability: string): string =>
  pluginTypes.find(type => type.id === capability)?.label ?? capability;

const categoryLabels: Record<string, string> = {
  utilities: 'Utilities',
  monitoring: 'Monitoring',
  system: 'System',
  appearance: 'Appearance',
  productivity: 'Productivity',
  media: 'Media',
  networking: 'Networking',
  audio: 'Audio',
};

const categoryLabel = (id: string): string =>
  categoryLabels[id] ?? id.charAt(0).toUpperCase() + id.slice(1);

const minCategorySize = 2;

function categoryOptions(plugins: Plugin[]): { id: string; label: string }[] {
  const counts: Record<string, number> = {};
  for (const plugin of plugins) {
    if (!plugin.category) continue;
    counts[plugin.category] = (counts[plugin.category] ?? 0) + 1;
  }
  const ids = Object.keys(counts)
    .filter(id => counts[id] >= minCategorySize)
    .sort((a, b) => counts[b] - counts[a] || a.localeCompare(b));
  return [{ id: 'all', label: 'All Plugins' }, ...ids.map(id => ({ id, label: categoryLabel(id) }))];
}

interface UrlFilters {
  tab: ActiveTab;
  type: string[];
  category: string[];
  compositor: string[];
}

type MultiFilterKey = Exclude<keyof UrlFilters, 'tab'>;

const multiFilterKeys: MultiFilterKey[] = ['type', 'category', 'compositor'];

const readListParam = (params: URLSearchParams, key: string): string[] =>
  (params.get(key) ?? '').split(',').filter(Boolean);

const readUrlFilters = (search: string): UrlFilters => {
  const params = new URLSearchParams(search);
  return {
    tab: params.get('tab') === 'themes' ? 'themes' : 'plugins',
    type: readListParam(params, 'type'),
    category: readListParam(params, 'category'),
    compositor: readListParam(params, 'compositor'),
  };
};

const writeUrlFilters = (search: string, filters: UrlFilters): string => {
  const params = new URLSearchParams(search);
  if (filters.tab === 'plugins') {
    params.delete('tab');
  } else {
    params.set('tab', filters.tab);
  }
  for (const key of multiFilterKeys) {
    if (filters[key].length === 0) {
      params.delete(key);
      continue;
    }
    params.set(key, filters[key].join(','));
  }
  return params.toString();
};

const toggleValue = (values: string[], value: string): string[] =>
  values.includes(value) ? values.filter(v => v !== value) : [...values, value];

interface FilterChipGroupProps {
  label: string;
  options: { id: string; label: string }[];
  selected: string[];
  onChange: (next: string[]) => void;
}

function FilterChipGroup({ label, options, selected, onChange }: FilterChipGroupProps) {
  return (
    <div className={styles.filterGroup}>
      <label className={styles.filterLabel}>{label}</label>
      <div className={styles.filterButtons}>
        {options.map(option => {
          const active = option.id === 'all' ? selected.length === 0 : selected.includes(option.id);
          return (
            <button
              key={option.id}
              className={`${styles.filterButton} ${active ? styles.active : ''}`}
              aria-pressed={active}
              onClick={() => onChange(option.id === 'all' ? [] : toggleValue(selected, option.id))}
            >
              {option.label}
            </button>
          );
        })}
      </div>
    </div>
  );
}

const compositors = [
  { id: 'all', label: 'All Compositors' },
  { id: 'niri', label: 'Niri' },
  { id: 'hyprland', label: 'Hyprland' },
  { id: 'sway', label: 'Sway' },
  { id: 'any', label: 'Any' },
];

const sortOptions = [
  { id: 'upvotes', label: 'Upvotes' },
  { id: 'updated_at', label: 'Last Updated' },
  { id: 'created_at', label: 'Date Added' },
  { id: 'name', label: 'Name' },
  { id: 'random', label: 'Random' },
];

// Sorts that only apply to plugins (themes have no upvotes or registry-added date).
const pluginOnlySorts = ['upvotes', 'created_at'];

const defaultSortOrder = (sort: string): 'asc' | 'desc' => (sort === 'name' ? 'asc' : 'desc');

const STATUS_LABELS: Record<string, string> = {
  broken: 'Broken',
  unmaintained: 'Unmaintained',
  deprecated: 'Deprecated',
  reviewed: 'Reviewed',
};

const STATUS_TOOLTIPS: Record<string, string> = {
  reviewed: 'Reviewed by catalog moderators for basic quality, ownership, and policy compliance. This is not a security guarantee.',
};

function wcagBadgeClass(level: string): string {
  switch (level) {
    case 'AAA':
      return styles.wcagAAA;
    case 'AA':
      return styles.wcagAA;
    default:
      return styles.wcagFail;
  }
}

const MODES = ['dark', 'light'] as const;

function wcagBreakdownRows(wcag: ThemeWCAG): ThemeWCAGBreakdown[] {
  const rows: ThemeWCAGBreakdown[] = [];
  for (const mode of MODES) {
    const result = wcag[mode];
    if (result?.breakdown) {
      rows.push(...result.breakdown);
    }
  }
  return rows;
}

// Rate every selectable config, not just the default, then advertise the level
// they all reach. "Partial" flags that some configs fall short, so the badge
// never over-promises for a theme whose flavors differ. Themes that only miss
// on accent colors still get credit for readable body text.
function wcagBadge(wcag: ThemeWCAG): { label: string; level: string } {
  const rows = wcagBreakdownRows(wcag);
  const tiers: Array<[keyof ThemeWCAGBreakdown, string]> = [
    ['level', ''],
    ['bodyLevel', ' body text'],
  ];

  for (const [key, suffix] of tiers) {
    const levels = rows.map(row => row[key] as string).filter(Boolean);
    const passing = levels.filter(level => level !== 'fail');
    if (!passing.length) {
      continue;
    }

    const level = passing.some(l => l === 'AA') ? 'AA' : 'AAA';
    const scope = passing.length === levels.length ? '' : ' (Partial)';
    return { label: `WCAG ${level}${suffix}${scope}`, level };
  }

  return { label: 'Below WCAG AA', level: 'fail' };
}

function wcagRowLabel(level: string): string {
  return level === 'fail' ? 'below AA' : `WCAG ${level}`;
}

function wcagTooltip(wcag: ThemeWCAG): string {
  const lines: string[] = ['Contrast measured per WCAG 2.2'];
  for (const row of wcagBreakdownRows(wcag)) {
    lines.push(`${row.name} (${row.mode}): ${wcagRowLabel(row.level)}`);
  }
  return lines.join('\n');
}

function statusBadgeClass(status: string): string {
  switch (status) {
    case 'broken':
      return styles.statusBroken;
    case 'unmaintained':
      return styles.statusUnmaintained;
    case 'deprecated':
      return styles.statusDeprecated;
    case 'reviewed':
      return styles.statusReviewed;
    default:
      return '';
  }
}

type ViewMode = 'grid' | 'list';

interface SortDropdownProps {
  id: string;
  options: { id: string; label: string }[];
  value: string;
  onChange: (value: string) => void;
}

function SortDropdown({ id, options, value, onChange }: SortDropdownProps) {
  const [open, setOpen] = useState(false);
  const containerRef = React.useRef<HTMLDivElement>(null);

  const selected = options.find(option => option.id === value);

  useEffect(() => {
    if (!open) {
      return;
    }

    const handlePointerDown = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setOpen(false);
      }
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setOpen(false);
      }
    };

    document.addEventListener('mousedown', handlePointerDown);
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('mousedown', handlePointerDown);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [open]);

  return (
    <div className={styles.sortDropdown} ref={containerRef}>
      <button
        type="button"
        id={id}
        className={`${styles.sortDropdownTrigger} ${open ? styles.sortDropdownTriggerOpen : ''}`}
        onClick={() => setOpen(current => !current)}
        aria-haspopup="listbox"
        aria-expanded={open}
      >
        <span>{selected?.label ?? 'Sort'}</span>
        <KeyboardArrowDown size={18} className={`${styles.sortDropdownChevron} ${open ? styles.sortDropdownChevronOpen : ''}`} />
      </button>
      {open && (
        <ul className={styles.sortDropdownMenu} role="listbox" aria-labelledby={id}>
          {options.map(option => (
            <li key={option.id} role="presentation">
              <button
                type="button"
                role="option"
                aria-selected={option.id === value}
                className={`${styles.sortDropdownOption} ${option.id === value ? styles.sortDropdownOptionActive : ''}`}
                onClick={() => {
                  onChange(option.id);
                  setOpen(false);
                }}
              >
                {option.label}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

export default function Plugins() {
  const location = useLocation();
  const history = useHistory();

  const initialFilters = readUrlFilters(location.search);

  const [activeTab, setActiveTab] = useState<ActiveTab>(initialFilters.tab);
  const [plugins, setPlugins] = useState<Plugin[]>([]);
  const [filteredPlugins, setFilteredPlugins] = useState<Plugin[]>([]);
  const [themes, setThemes] = useState<Theme[]>([]);
  const [filteredThemes, setFilteredThemes] = useState<Theme[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedCategories, setSelectedCategories] = useState<string[]>(initialFilters.category);
  const [selectedTypes, setSelectedTypes] = useState<string[]>(initialFilters.type);
  const [selectedCompositors, setSelectedCompositors] = useState<string[]>(initialFilters.compositor);
  const [searchQuery, setSearchQuery] = useState('');
  const [showFirstPartyOnly, setShowFirstPartyOnly] = useState(false);
  const [showReviewedOnly, setShowReviewedOnly] = useState(false);
  const [hideBroken, setHideBroken] = useState(false);
  const [sortBy, setSortBy] = useState('upvotes');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [lightboxPlugin, setLightboxPlugin] = useState<Plugin | null>(null);
  const [selectedVariants, setSelectedVariants] = useState<Record<string, string>>({});
  const [selectedFlavors, setSelectedFlavors] = useState<Record<string, string>>({});
  const [selectedAccents, setSelectedAccents] = useState<Record<string, string>>({});
  const [viewMode, setViewMode] = useState<ViewMode>('grid');

  const pluginsById = React.useMemo(() => {
    const map: Record<string, Plugin> = {};
    for (const p of plugins) {
      map[p.id] = p;
    }
    return map;
  }, [plugins]);

  const categories = React.useMemo(() => categoryOptions(plugins), [plugins]);

  useEffect(() => {
    const next = writeUrlFilters(location.search, {
      tab: activeTab,
      type: selectedTypes,
      category: selectedCategories,
      compositor: selectedCompositors,
    });
    if (next !== new URLSearchParams(location.search).toString()) {
      history.replace({ search: next });
    }
  }, [activeTab, selectedTypes, selectedCategories, selectedCompositors]);

  useEffect(() => {
    const filters = readUrlFilters(location.search);
    setActiveTab(filters.tab);
    setSelectedTypes(filters.type);
    setSelectedCategories(filters.category);
    setSelectedCompositors(filters.compositor);
  }, [location.search]);

  useEffect(() => {
    switch (activeTab) {
      case 'themes':
        fetchThemes();
        break;
      default:
        fetchPlugins();
    }
  }, [sortBy, sortOrder, activeTab]);

  useEffect(() => {
    filterPlugins();
  }, [plugins, selectedCategories, selectedTypes, selectedCompositors, searchQuery, showFirstPartyOnly, showReviewedOnly, hideBroken]);

  useEffect(() => {
    filterThemes();
  }, [themes, searchQuery]);

  const fetchPlugins = async () => {
    try {
      setLoading(true);
      const isDev = process.env.NODE_ENV === 'development';
      const apiOrigin = isDev ? 'http://localhost:8337' : 'https://api.duxlabs.dev';
      const apiUrl = `${apiOrigin}/plugins?sortBy=${sortBy}&order=${sortOrder}`;
      const response = await fetch(apiUrl);
      if (!response.ok) {
        throw new Error('Failed to fetch plugins');
      }
      const data: PluginsResponse = await response.json();
      const pluginList = (data.plugins ?? []).map(p => ({
        ...p,
        previewUrl: p.previewUrl ? `${apiOrigin}/previews/${p.id}` : p.previewUrl,
      }));
      setPlugins(pluginList);
      setFilteredPlugins(pluginList);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const filterPlugins = () => {
    let filtered = [...(plugins ?? [])];

    if (selectedCategories.length > 0) {
      filtered = filtered.filter(p => selectedCategories.includes(p.category));
    }

    if (selectedTypes.length > 0) {
      filtered = filtered.filter(p => selectedTypes.some(type => p.capabilities.includes(type)));
    }

    if (selectedCompositors.length > 0) {
      filtered = filtered.filter(p =>
        p.compositors.includes('any') || selectedCompositors.some(comp => p.compositors.includes(comp))
      );
    }

    if (showFirstPartyOnly) {
      filtered = filtered.filter(p => p.firstParty);
    }

    if (showReviewedOnly) {
      filtered = filtered.filter(p => p.status?.includes('reviewed'));
    }

    if (hideBroken) {
      filtered = filtered.filter(p =>
        !p.status?.includes('broken') && !p.status?.includes('deprecated')
      );
    }

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(p =>
        p.name.toLowerCase().includes(query) ||
        p.description.toLowerCase().includes(query) ||
        p.author.toLowerCase().includes(query) ||
        p.id.toLowerCase().includes(query)
      );
    }

    setFilteredPlugins(filtered);
  };

  const fetchThemes = async () => {
    try {
      setLoading(true);
      const isDev = process.env.NODE_ENV === 'development';
      const baseUrl = isDev ? 'http://localhost:8337/themes' : 'https://api.duxlabs.dev/themes';
      const themeSort = pluginOnlySorts.includes(sortBy) ? 'updated_at' : sortBy;
      const apiUrl = `${baseUrl}?sortBy=${themeSort}`;
      const response = await fetch(apiUrl);
      if (!response.ok) {
        throw new Error('Failed to fetch themes');
      }
      const data: ThemesResponse = await response.json();
      const themeList = data.themes ?? [];
      setThemes(themeList);
      setFilteredThemes(themeList);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const filterThemes = () => {
    if (!searchQuery) {
      setFilteredThemes(themes);
      return;
    }

    const query = searchQuery.toLowerCase();
    const filtered = themes.filter(t =>
      t.name.toLowerCase().includes(query) ||
      t.description.toLowerCase().includes(query) ||
      t.author.toLowerCase().includes(query) ||
      t.id.toLowerCase().includes(query)
    );
    setFilteredThemes(filtered);
  };

  useEffect(() => {
    if (!lightboxPlugin) return;
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setLightboxPlugin(null);
    };
    document.addEventListener('keydown', handleKey);
    return () => document.removeEventListener('keydown', handleKey);
  }, [lightboxPlugin]);

  const getPluginIcon = (plugin: Plugin) => {
    if (plugin.icon) {
      return (
        <span className={`material-symbols-rounded ${styles.pluginIconMaterial}`}>
          {plugin.icon}
        </span>
      );
    }
    return null;
  };

  const formatUpdatedAt = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInMs = now.getTime() - date.getTime();
    const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24));

    if (diffInDays === 0) {
      return 'Updated today';
    }
    if (diffInDays === 1) {
      return 'Updated yesterday';
    }
    if (diffInDays < 7) {
      return `Updated ${diffInDays} days ago`;
    }
    if (diffInDays < 30) {
      const weeks = Math.floor(diffInDays / 7);
      return `Updated ${weeks} ${weeks === 1 ? 'week' : 'weeks'} ago`;
    }
    if (diffInDays < 365) {
      const months = Math.floor(diffInDays / 30);
      return `Updated ${months} ${months === 1 ? 'month' : 'months'} ago`;
    }
    const years = Math.floor(diffInDays / 365);
    return `Updated ${years} ${years === 1 ? 'year' : 'years'} ago`;
  };

  const getSelectedVariant = (theme: Theme): string => {
    if (!theme.variants || theme.variants.type === 'multi') return '';
    return selectedVariants[theme.id] || theme.variants.default || '';
  };

  const getSelectedFlavor = (theme: Theme): string => {
    if (!theme.variants?.type || theme.variants.type !== 'multi') return '';
    return selectedFlavors[theme.id] || theme.variants.defaults?.dark?.flavor || '';
  };

  const getFlavorMode = (theme: Theme): 'dark' | 'light' => {
    const flavorId = getSelectedFlavor(theme);
    const flavor = theme.variants?.flavors?.find(f => f.id === flavorId);
    return flavor?.light ? 'light' : 'dark';
  };

  const getSelectedAccent = (theme: Theme): string => {
    if (!theme.variants?.type || theme.variants.type !== 'multi') return '';
    return selectedAccents[theme.id] || theme.variants.defaults?.dark?.accent || '';
  };

  const getResolvedColors = (theme: Theme, mode: 'dark' | 'light'): Record<string, string> => {
    const baseColors = theme[mode] || {};
    if (!theme.variants) return baseColors;

    if (theme.variants.type === 'multi') {
      const flavorId = getSelectedFlavor(theme);
      const accentId = getSelectedAccent(theme);
      const flavor = theme.variants.flavors?.find(f => f.id === flavorId);
      const accent = theme.variants.accents?.find(a => a.id === accentId);

      const flavorMode = flavor?.light ? 'light' : 'dark';
      if (flavorMode !== mode) return baseColors;

      const flavorColors = flavor?.[mode] || {};
      const accentColors = (accent?.[flavorId] as Record<string, string>) || {};

      return { ...baseColors, ...flavorColors, ...accentColors };
    }

    const variantId = getSelectedVariant(theme);
    const variant = theme.variants.options?.find(v => v.id === variantId);
    if (!variant) return baseColors;

    return { ...baseColors, ...variant[mode] };
  };

  const getPreviewUrl = (theme: Theme): string => {
    if (!theme.variants) return theme.previewUrl;

    if (theme.variants.type === 'multi') {
      const flavorId = getSelectedFlavor(theme);
      const accentId = getSelectedAccent(theme);
      return theme.previewUrl.replace('preview.svg', `preview-${flavorId}-${accentId}.svg`);
    }

    const variantId = getSelectedVariant(theme);
    return theme.previewUrl.replace('preview.svg', `preview-${variantId}.svg`);
  };

  const getInstallUrl = (theme: Theme): string => {
    const base = `dms://theme/install/${theme.id}`;
    if (!theme.variants) return base;

    if (theme.variants.type === 'multi') {
      const flavor = getSelectedFlavor(theme);
      const accent = getSelectedAccent(theme);
      return `${base}?flavor=${flavor}&accent=${accent}`;
    }

    return `${base}?variant=${getSelectedVariant(theme)}`;
  };

  return (
    <Layout
      title="Plugins & Themes"
      description="Discover and install plugins and themes for Dux Shell - extend your desktop with widgets, launchers, and more.">
      <Head>
        <link
          rel="stylesheet"
          href="https://fonts.googleapis.com/css2?family=Material+Symbols+Rounded:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200"
        />
      </Head>
      <div className={styles.container}>
        <LandingBackground />
        <div className={styles.content}>
          <section className={styles.header}>
            <h1 className={styles.title}>
              Explore <span className={styles.gradientText}>{activeTab === 'plugins' ? 'Plugins' : 'Themes'}</span>
            </h1>
            <p className={styles.subtitle}>
              {activeTab === 'plugins'
                ? 'Extend Dux Shell with powerful plugins for widgets, launchers, and more'
                : 'Customize your desktop with beautiful color schemes'}
            </p>
            <Segmented
              label="Catalog"
              className={styles.tabContainer}
              value={activeTab}
              onChange={setActiveTab}
              options={[
                {value: 'plugins', label: 'Plugins'},
                {value: 'themes', label: 'Themes'},
              ]}
            />
          </section>

          <section className={styles.filters}>
            <div className={styles.topControls}>
              <input
                type="text"
                placeholder={activeTab === 'plugins' ? 'Search plugins...' : 'Search themes...'}
                aria-label={activeTab === 'plugins' ? 'Search plugins' : 'Search themes'}
                className={styles.searchInput}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
              <div className={styles.controlsRight}>
                {activeTab === 'plugins' && (
                  <label className={styles.checkboxLabel}>
                    <input
                      type="checkbox"
                      checked={showFirstPartyOnly}
                      onChange={(e) => setShowFirstPartyOnly(e.target.checked)}
                      className={styles.checkbox}
                    />
                    <span>First-party only</span>
                  </label>
                )}
                {activeTab === 'plugins' && (
                  <label className={styles.checkboxLabel}>
                    <input
                      type="checkbox"
                      checked={showReviewedOnly}
                      onChange={(e) => setShowReviewedOnly(e.target.checked)}
                      className={styles.checkbox}
                    />
                    <span>Reviewed only</span>
                  </label>
                )}
                {activeTab === 'plugins' && (
                  <label className={styles.checkboxLabel}>
                    <input
                      type="checkbox"
                      checked={hideBroken}
                      onChange={(e) => setHideBroken(e.target.checked)}
                      className={styles.checkbox}
                    />
                    <span>Hide broken</span>
                  </label>
                )}
                <div className={styles.sortGroup}>
                  <span className={styles.filterLabel} id="sort-select-label">Sort by</span>
                  <SortDropdown
                    id="sort-select"
                    options={sortOptions.filter(option => activeTab === 'plugins' || !pluginOnlySorts.includes(option.id))}
                    value={activeTab === 'themes' && pluginOnlySorts.includes(sortBy) ? 'updated_at' : sortBy}
                    onChange={(next) => {
                      setSortBy(next);
                      setSortOrder(defaultSortOrder(next));
                    }}
                  />
                  {activeTab === 'plugins' && sortBy !== 'random' && (
                    <IconButton
                      label={sortOrder === 'asc' ? 'Sort ascending' : 'Sort descending'}
                      title={sortOrder === 'asc' ? 'Ascending' : 'Descending'}
                      onClick={() => setSortOrder(prev => (prev === 'asc' ? 'desc' : 'asc'))}
                    >
                      {sortOrder === 'asc' ? <ArrowUpward /> : <ArrowDownward />}
                    </IconButton>
                  )}
                </div>
                <Segmented
                  label="View"
                  className={styles.viewToggle}
                  value={viewMode}
                  onChange={setViewMode}
                  options={[
                    {value: 'grid', icon: <GridView />, label: <span className="sr-only">Grid view</span>},
                    {value: 'list', icon: <ViewList />, label: <span className="sr-only">List view</span>},
                  ]}
                />
              </div>
            </div>

            {activeTab === 'plugins' && (
              <div className={styles.filterRow}>
                <FilterChipGroup label="Category" options={categories} selected={selectedCategories} onChange={setSelectedCategories} />
                <FilterChipGroup label="Type" options={pluginTypes} selected={selectedTypes} onChange={setSelectedTypes} />
                <FilterChipGroup label="Compositor" options={compositors} selected={selectedCompositors} onChange={setSelectedCompositors} />
              </div>
            )}

            <div className={styles.resultsCount}>
              {activeTab === 'plugins'
                ? `${filteredPlugins.length} ${filteredPlugins.length === 1 ? 'plugin' : 'plugins'} found`
                : `${filteredThemes.length} ${filteredThemes.length === 1 ? 'theme' : 'themes'} found`}
            </div>
          </section>

          {loading && (
            <div className={styles.loadingContainer}>
              <div className={styles.spinner}></div>
              <p>{`Loading ${activeTab}...`}</p>
            </div>
          )}

          {error && (
            <div className={styles.errorContainer}>
              <p>Error: {error}</p>
              <button
                onClick={activeTab === 'plugins' ? fetchPlugins : fetchThemes}
                className={styles.retryButton}
              >
                Retry
              </button>
            </div>
          )}

          {!loading && !error && activeTab === 'plugins' && (
            <section className={viewMode === 'grid' ? styles.pluginsGrid : styles.pluginsList}>
              {filteredPlugins.map(plugin => (
                <div key={plugin.id} className={`${styles.pluginCard} ${viewMode === 'list' ? styles.listCard : ''}`}>
                  {(plugin.previewUrl || plugin.screenshot) && viewMode === 'grid' && (
                    <div
                      className={styles.pluginImage}
                      onClick={() => setLightboxPlugin(plugin)}
                      role="button"
                      title="Click to expand"
                    >
                      <img src={plugin.previewUrl || plugin.screenshot} alt={plugin.name} loading="lazy" />
                    </div>
                  )}
                  <div className={styles.pluginContent}>
                    <div className={styles.pluginHeader}>
                      {getPluginIcon(plugin)}
                      <div className={styles.pluginTitleGroup}>
                        <h3 className={styles.pluginName}>
                          {plugin.name}
                          {plugin.firstParty && (
                            <span className={styles.officialBadge}>Official</span>
                          )}
                        </h3>
                        <p className={styles.pluginAuthor}>by {plugin.author}</p>
                      </div>
                      {plugin.issueUrl && (
                        <a
                          href={plugin.issueUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className={styles.upvoteChip}
                          title="Upvote on GitHub"
                        >
                          <ThumbUp size={15} />
                          {plugin.upvotes}
                        </a>
                      )}
                    </div>

                    <p className={styles.pluginDescription}>{plugin.description}</p>

                    <div className={styles.pluginMeta}>
                      <div className={styles.pluginTags}>
                        {plugin.status?.map(s => (
                          <span key={s} className={`${styles.statusBadge} ${statusBadgeClass(s)}`} title={STATUS_TOOLTIPS[s]}>
                            {STATUS_LABELS[s] || s}
                          </span>
                        ))}
                        <span className={styles.tag}>{categoryLabel(plugin.category)}</span>
                        {viewMode === 'grid' && plugin.version && (
                          <span className={styles.tag}>v{plugin.version}</span>
                        )}
                        {viewMode === 'grid' && plugin.updated_at && (
                          <span className={styles.tag}>{formatUpdatedAt(plugin.updated_at)}</span>
                        )}
                      </div>

                      <div className={styles.pluginCapabilities}>
                        {plugin.capabilities.slice(0, 2).map(cap => (
                          <span key={cap} className={styles.capability}>
                            {pluginTypeLabel(cap)}
                          </span>
                        ))}
                        {plugin.capabilities.length > 2 && (
                          <span className={styles.capability}>
                            +{plugin.capabilities.length - 2}
                          </span>
                        )}
                      </div>

                      {viewMode === 'grid' && plugin.compositors.length > 0 && (
                        <div className={styles.pluginCompositors}>
                          {plugin.compositors.map(comp => (
                            <span key={comp} className={styles.compositor}>
                              {comp}
                            </span>
                          ))}
                        </div>
                      )}

                      {viewMode === 'grid' && plugin.dependencies.length > 0 && (
                        <div className={styles.pluginDeps}>
                          <strong>Dependencies:</strong> {plugin.dependencies.join(', ')}
                        </div>
                      )}

                      {viewMode === 'grid' && plugin.requires_dms && (
                        <div className={styles.pluginRequires}>
                          <strong>Requires DMS:</strong> {plugin.requires_dms}+
                        </div>
                      )}

                      {plugin.similar && plugin.similar.length > 0 && (
                        <div className={styles.pluginRelated}>
                          <strong>Related:</strong>
                          {plugin.similar.map(id => (
                            <button
                              key={id}
                              type="button"
                              className={styles.relatedChip}
                              onClick={() => {
                                setSearchQuery(pluginsById[id]?.name || id);
                              }}
                            >
                              {pluginsById[id]?.name || id}
                            </button>
                          ))}
                        </div>
                      )}
                    </div>

                    <div className={styles.pluginActions}>
                      <a
                        href={`dms://plugin/install/${plugin.id}`}
                        className={styles.installButton}
                      >
                        <Download size={18} />
                        {viewMode === 'grid' && 'Install'}
                      </a>
                      <a
                        href={plugin.repo}
                        target="_blank"
                        rel="noopener noreferrer"
                        className={styles.repoButton}
                        title="View Repository"
                      >
                        <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                          <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
                        </svg>
                        {viewMode === 'grid' && <span>Repository</span>}
                      </a>
                      {plugin.issueUrl && (
                        <a
                          href={plugin.issueUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className={styles.discussButton}
                          title="Discuss / Vote"
                        >
                          <Forum size={18} />
                          {viewMode === 'grid' && <span>Discuss</span>}
                        </a>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </section>
          )}

          {!loading && !error && activeTab === 'themes' && (
            <section className={styles.pluginsGrid}>
              {filteredThemes.map(theme => (
                <div key={theme.id} className={styles.pluginCard}>
                  <div className={styles.themePreview}>
                    <img src={getPreviewUrl(theme)} alt={theme.name} loading="lazy" />
                  </div>
                  <div className={styles.pluginContent}>
                    <div className={styles.pluginHeader}>
                      <div className={styles.pluginTitleGroup}>
                        <h3 className={styles.pluginName}>{theme.name}</h3>
                        <p className={styles.pluginAuthor}>by {theme.author}</p>
                      </div>
                    </div>

                    <p className={styles.pluginDescription}>{theme.description}</p>

                    {theme.variants && theme.variants.type !== 'multi' && theme.variants.options && (
                      <div className={styles.variantSelector}>
                        <label className={styles.variantLabel}>Variant</label>
                        <div className={styles.variantButtons}>
                          {theme.variants.options.map(variant => (
                            <button
                              key={variant.id}
                              className={`${styles.variantButton} ${getSelectedVariant(theme) === variant.id ? styles.active : ''}`}
                              aria-pressed={getSelectedVariant(theme) === variant.id}
                              onClick={() => setSelectedVariants(prev => ({ ...prev, [theme.id]: variant.id }))}
                            >
                              {variant.name}
                            </button>
                          ))}
                        </div>
                      </div>
                    )}

                    {theme.variants?.type === 'multi' && (
                      <>
                        <div className={styles.variantSelector}>
                          <label className={styles.variantLabel}>Flavor</label>
                          <div className={styles.variantButtons}>
                            {theme.variants.flavors?.map(flavor => (
                              <button
                                key={flavor.id}
                                className={`${styles.variantButton} ${getSelectedFlavor(theme) === flavor.id ? styles.active : ''}`}
                                aria-pressed={getSelectedFlavor(theme) === flavor.id}
                                onClick={() => setSelectedFlavors(prev => ({ ...prev, [theme.id]: flavor.id }))}
                              >
                                {flavor.name}
                              </button>
                            ))}
                          </div>
                        </div>
                        <div className={styles.variantSelector}>
                          <label className={styles.variantLabel}>Accent</label>
                          <div className={styles.variantButtons}>
                            {theme.variants.accents?.map(accent => (
                              <button
                                key={accent.id}
                                className={`${styles.variantButton} ${getSelectedAccent(theme) === accent.id ? styles.active : ''}`}
                                aria-pressed={getSelectedAccent(theme) === accent.id}
                                onClick={() => setSelectedAccents(prev => ({ ...prev, [theme.id]: accent.id }))}
                              >
                                {accent.name}
                              </button>
                            ))}
                          </div>
                        </div>
                      </>
                    )}

                    <div className={styles.themeColors}>
                      {(theme.variants?.type !== 'multi' || getFlavorMode(theme) === 'dark') && (
                        <div className={styles.colorScheme}>
                          <span className={styles.schemeLabel}>{theme.variants?.type === 'multi' ? getSelectedFlavor(theme) : 'Dark'}</span>
                          <div className={styles.colorSwatches}>
                            <div
                              className={styles.colorSwatch}
                              style={{ backgroundColor: getResolvedColors(theme, 'dark').primary }}
                              title="Primary"
                            />
                            <div
                              className={styles.colorSwatch}
                              style={{ backgroundColor: getResolvedColors(theme, 'dark').secondary }}
                              title="Secondary"
                            />
                            <div
                              className={styles.colorSwatch}
                              style={{ backgroundColor: getResolvedColors(theme, 'dark').surface }}
                              title="Surface"
                            />
                            <div
                              className={styles.colorSwatch}
                              style={{ backgroundColor: getResolvedColors(theme, 'dark').background }}
                              title="Background"
                            />
                          </div>
                        </div>
                      )}
                      {(theme.variants?.type !== 'multi' || getFlavorMode(theme) === 'light') && (
                        <div className={styles.colorScheme}>
                          <span className={styles.schemeLabel}>{theme.variants?.type === 'multi' ? getSelectedFlavor(theme) : 'Light'}</span>
                          <div className={styles.colorSwatches}>
                            <div
                              className={styles.colorSwatch}
                              style={{ backgroundColor: getResolvedColors(theme, 'light').primary }}
                              title="Primary"
                            />
                            <div
                              className={styles.colorSwatch}
                              style={{ backgroundColor: getResolvedColors(theme, 'light').secondary }}
                              title="Secondary"
                            />
                            <div
                              className={styles.colorSwatch}
                              style={{ backgroundColor: getResolvedColors(theme, 'light').surface }}
                              title="Surface"
                            />
                            <div
                              className={styles.colorSwatch}
                              style={{ backgroundColor: getResolvedColors(theme, 'light').background }}
                              title="Background"
                            />
                          </div>
                        </div>
                      )}
                    </div>

                    <div className={styles.pluginTags}>
                      {theme.wcag && (
                        <span
                          className={`${styles.statusBadge} ${wcagBadgeClass(wcagBadge(theme.wcag).level)}`}
                          title={wcagTooltip(theme.wcag)}
                        >
                          {wcagBadge(theme.wcag).label}
                        </span>
                      )}
                      {theme.version && (
                        <span className={styles.tag}>v{theme.version}</span>
                      )}
                      {theme.updated_at && (
                        <span className={styles.tag}>{formatUpdatedAt(theme.updated_at)}</span>
                      )}
                    </div>

                    <div className={styles.pluginActions}>
                      <a
                        href={getInstallUrl(theme)}
                        className={styles.installButton}
                      >
                        <Download size={18} />
                        Install Theme
                      </a>
                    </div>
                  </div>
                </div>
              ))}
            </section>
          )}

          {!loading && !error && activeTab === 'plugins' && filteredPlugins.length === 0 && (
            <div className={styles.emptyState}>
              <p>No plugins found matching your criteria.</p>
              <button
                onClick={() => {
                  setSelectedCategories([]);
                  setSelectedTypes([]);
                  setSelectedCompositors([]);
                  setSearchQuery('');
                  setShowFirstPartyOnly(false);
                  setShowReviewedOnly(false);
                  setHideBroken(false);
                  setSortBy('upvotes');
                  setSortOrder('desc');
                }}
                className={styles.resetButton}
              >
                Reset Filters
              </button>
            </div>
          )}

          {!loading && !error && activeTab === 'themes' && filteredThemes.length === 0 && (
            <div className={styles.emptyState}>
              <p>No themes found matching your criteria.</p>
              <button
                onClick={() => {
                  setSearchQuery('');
                  setSortBy('updated_at');
                }}
                className={styles.resetButton}
              >
                Reset Filters
              </button>
            </div>
          )}
        </div>
      </div>
      {lightboxPlugin && (
        <div className={styles.lightboxOverlay} onClick={() => setLightboxPlugin(null)}>
          <div className={styles.lightboxContent} onClick={e => e.stopPropagation()}>
            <img
              src={lightboxPlugin.previewUrl || lightboxPlugin.screenshot}
              alt={lightboxPlugin.name}
              className={styles.lightboxImage}
            />
            <div className={styles.lightboxCaption}>
              <span>{lightboxPlugin.name}</span>
              {lightboxPlugin.screenshot && (
                <a href={lightboxPlugin.screenshot} target="_blank" rel="noopener noreferrer">
                  Open raw screenshot
                </a>
              )}
              <button onClick={() => setLightboxPlugin(null)} aria-label="Close">
                ✕
              </button>
            </div>
          </div>
        </div>
      )}
    </Layout>
  );
}
