import { themes as prismThemes } from 'prism-react-renderer';
import type { Config } from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';
import type { PrismTheme } from 'prism-react-renderer';
import path from 'path';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

// Custom Dank Purple Prism Theme
const dankPurple: PrismTheme = {
  plain: {
    color: '#e4e4e7',
    backgroundColor: '#2e2a3d',
  },
  styles: [
    {
      types: ['comment', 'prolog', 'doctype', 'cdata'],
      style: {
        color: '#7c7c99',
        fontStyle: 'italic',
      },
    },
    {
      types: ['punctuation'],
      style: {
        color: '#c0c0d0',
      },
    },
    {
      types: ['namespace'],
      style: {
        opacity: 0.7,
      },
    },
    {
      types: ['property', 'tag', 'constant', 'symbol', 'deleted'],
      style: {
        color: '#f78c6c',
      },
    },
    {
      types: ['boolean', 'number'],
      style: {
        color: '#f78c6c',
      },
    },
    {
      types: ['selector', 'attr-name', 'string', 'char', 'builtin', 'inserted'],
      style: {
        color: '#c3e88d',
      },
    },
    {
      types: ['operator', 'entity', 'url', 'variable'],
      style: {
        color: '#89ddff',
      },
    },
    {
      types: ['atrule', 'attr-value', 'keyword'],
      style: {
        color: '#c792ea',
      },
    },
    {
      types: ['function', 'class-name'],
      style: {
        color: '#82aaff',
      },
    },
    {
      types: ['regex', 'important'],
      style: {
        color: '#f07178',
      },
    },
    {
      types: ['important', 'bold'],
      style: {
        fontWeight: 'bold',
      },
    },
    {
      types: ['italic'],
      style: {
        fontStyle: 'italic',
      },
    },
  ],
};

const dankPurpleLight: PrismTheme = {
  plain: {
    color: '#2e2a3d',
    backgroundColor: '#f8f7fb',
  },
  styles: [
    {
      types: ['comment', 'prolog', 'doctype', 'cdata'],
      style: {
        color: '#6e7c93',
        fontStyle: 'italic',
      },
    },
    {
      types: ['punctuation'],
      style: {
        color: '#5a5a77',
      },
    },
    {
      types: ['namespace'],
      style: {
        opacity: 0.7,
      },
    },
    {
      types: ['property', 'tag', 'constant', 'symbol', 'deleted'],
      style: {
        color: '#d65d0e',
      },
    },
    {
      types: ['boolean', 'number'],
      style: {
        color: '#d65d0e',
      },
    },
    {
      types: ['selector', 'attr-name', 'string', 'char', 'builtin', 'inserted'],
      style: {
        color: '#427b58',
      },
    },
    {
      types: ['operator', 'entity', 'url', 'variable'],
      style: {
        color: '#076678',
      },
    },
    {
      types: ['atrule', 'attr-value', 'keyword'],
      style: {
        color: '#8f3f71',
      },
    },
    {
      types: ['function', 'class-name'],
      style: {
        color: '#0969da',
      },
    },
    {
      types: ['regex', 'important'],
      style: {
        color: '#cc241d',
      },
    },
    {
      types: ['important', 'bold'],
      style: {
        fontWeight: 'bold',
      },
    },
    {
      types: ['italic'],
      style: {
        fontStyle: 'italic',
      },
    },
  ],
};

const proseOnly = (content: string) =>
  content
    .replace(/^export const[\s\S]*?^};$/gm, '')
    .replace(/^import .*$/gm, '')
    .replace(/\{\/\*[\s\S]*?\*\/\}/g, '')
    .replace(/<[^>]*>/g, '')
    .replace(/^\s*(light|dark|sources|style|alt)[:=].*$/gm, '')
    .replace(/^\|.*$/gm, '')
    .replace(/```[\s\S]*?```/g, '');

const config: Config = {
  title: 'Dux',
  tagline: 'Dux Modular Shell and tools: a modern Wayland desktop suite for niri, Hyprland, mangowm, dwl, labwc, and Sway.',
  favicon: 'img/favicon.ico',

  // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
  future: {
    v4: true, // Improve compatibility with the upcoming Docusaurus v4
    faster: {
      // SWC HTML minifier strips optional <main>/<div> tags on blog pages and
      // breaks the three-column layout (content appears unformatted below footer).
      swcHtmlMinimizer: false,
    },
  },

  // Add favicon, icons, and Open Graph meta tags
  headTags: [
    {
      tagName: 'link',
      attributes: {
        rel: 'icon',
        type: 'image/png',
        sizes: '16x16',
        href: '/img/favicon-16x16.png',
      },
    },
    {
      tagName: 'link',
      attributes: {
        rel: 'icon',
        type: 'image/png',
        sizes: '32x32',
        href: '/img/favicon-32x32.png',
      },
    },
    {
      tagName: 'link',
      attributes: {
        rel: 'icon',
        type: 'image/png',
        sizes: '48x48',
        href: '/img/favicon-48x48.png',
      },
    },
    {
      tagName: 'link',
      attributes: {
        rel: 'apple-touch-icon',
        sizes: '180x180',
        href: '/img/apple-touch-icon.png',
      },
    },
    {
      tagName: 'link',
      attributes: {
        rel: 'manifest',
        href: '/img/site.webmanifest',
      },
    },
    // Open Graph meta tags for social media previews (fallback - per-page images override this)
    {
      tagName: 'meta',
      attributes: {
        property: 'og:type',
        content: 'website',
      },
    },
    {
      tagName: 'meta',
      attributes: {
        property: 'og:site_name',
        content: 'Dux',
      },
    },
    // Twitter Card meta tags
    {
      tagName: 'meta',
      attributes: {
        name: 'twitter:card',
        content: 'summary_large_image',
      },
    },
  ],

  // Set the production url of your site here
  url: 'https://duxlabs.dev',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'duxlabs-dev', // Usually your GitHub org/user name.
  projectName: 'dux-docs', // Usually your repo name.

  onBrokenLinks: 'throw',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/duxlabs-dev/dux-docs/tree/master',
          lastVersion: '2.0',
          versions: {
            current: {
              label: '2.1',
              path: '2.1',
              banner: 'unreleased',
            },
            '2.0': {
              label: '2.0',
              path: '',
            },
            '1.6': {
              label: '1.6',
              path: '1.6',
            },
          },
        },
        blog: {
          showReadingTime: true,
          readingTime: ({content, locale, defaultReadingTime}) =>
            defaultReadingTime({content: proseOnly(content), locale}),
          blogTitle: 'Dux Updates',
          blogDescription: 'News, releases, and updates from the Dux project.',
          blogSidebarTitle: 'Recent Posts',
          blogSidebarCount: 5,
          feedOptions: {
            type: ['rss', 'atom'],
            copyright: `Copyright © ${new Date().getFullYear()} Dux Labs`,
          },
        },
        theme: {
          customCss: ['./src/css/m3e.css', './src/css/custom.css'],
        },
      } satisfies Preset.Options,
    ],
  ],

  markdown: {
    mermaid: true,
  },
  themes: ['@docusaurus/theme-mermaid'],

  plugins: [
    () => ({
      name: 'm3e',
      configurePostCss(options) {
        options.plugins.push(require('@tailwindcss/postcss'));
        return options;
      },
      configureWebpack() {
        return {resolve: {alias: {'@': path.resolve(__dirname, 'src')}}};
      },
    }),
    [
      require.resolve('@easyops-cn/docusaurus-search-local'),
      {
        hashed: true,
        language: ['en'],
        highlightSearchTermsOnTargetPage: true,
        explicitSearchResultPath: true,
      },
    ],
    [
      require.resolve('plugin-image-zoom'),
      {
        id: 'image-zoom-fullscreen',
      },
    ],
    [
      require.resolve('@actinc/docusaurus-plugin-panzoom'),
      {
        id: 'mermaid-panzoom',
      },
    ],
    [
      '@dipakparmar/docusaurus-plugin-umami',
      {
        websiteID: 'b79205dd-087a-4fbd-b7d7-c60650c76414',
        analyticsDomain: 'umami.duxlabs.dev',
        dataAutoTrack: true,
        dataDomains: 'duxlabs.dev,danklinux.com',
      },
    ],
    [
      '@docusaurus/plugin-client-redirects',
      {
        redirects: [{from: '/blog/v1-7-release', to: '/blog/v2-0-release'}],
        createRedirects(existingPath: string) {
          const map: [string, string[]][] = [
            ['/docs/dms/', ['/docs/dankmaterialshell/']],
            ['/docs/dms-greeter/', ['/docs/dankgreeter/']],
            ['/docs/dsearch/', ['/docs/danksearch/']],
            ['/docs/dcal/', ['/docs/dankcalendar/']],
            ['/docs/repo/', ['/docs/packaging/', '/docs/danklinux/']],
            ['/docs/dms', ['/docs/dankmaterialshell']],
            ['/docs/dms-greeter', ['/docs/dankgreeter']],
            ['/docs/dsearch', ['/docs/danksearch']],
            ['/docs/dcal', ['/docs/dankcalendar']],
            ['/docs/repo', ['/docs/packaging', '/docs/danklinux']],
            ['/docs/duxinstall', ['/docs/dankinstall']],
          ];
          for (const [neu, olds] of map) {
            if (existingPath === neu || existingPath.startsWith(neu.endsWith('/') ? neu : neu + '/')) {
              return olds.map((old) => existingPath.replace(neu, old));
            }
          }
          return undefined;
        },
      },
    ],
    [
      '@acid-info/docusaurus-og',
      {
        path: './preview-images',
        imageRenderers: {
          'docusaurus-plugin-content-docs': require('./lib/ImageRenderers').docs,
          'docusaurus-plugin-content-pages': require('./lib/ImageRenderers').pages,
          'docusaurus-plugin-content-blog': require('./lib/ImageRenderers').blog,
        },
      },
    ],
  ],

  themeConfig: {
    // Replace with your project's social card (fallback for pages without images)
    image: 'img/homepage/dux-preview.png',
    announcementBar: {
      id: 'v2-1-release',
      content:
        '<b>DMS 2.1 Stormborn.</b> <span class="announcement-detail">New system monitor, video wallpapers, revamped media player, interface density, numerous dux files enhancements, and more.</span> <a class="announcement-cta" href="/blog/v2-1-release">Read the announcement</a>',
      backgroundColor: '#6B46C1',
      textColor: '#ffffff',
      isCloseable: true,
    },
    colorMode: {
      defaultMode: 'dark',
      respectPrefersColorScheme: true,
    },
    docs: {
      sidebar: {
        hideable: true,
        autoCollapseCategories: true,
      },
    },
    // Panzoom configuration for Mermaid diagrams only (images use fullscreen zoom)
    zoom: {
      selectors: [
        'div.mermaid[data-processed="true"]',
        'div.docusaurus-mermaid-container',
      ],
      wrap: true,
      timeout: 1000,
    },
    navbar: {
      logo: {
        alt: 'Dux Labs Logo',
        src: 'img/dux-labs-logo-light.svg',
        srcDark: 'img/dux-labs-logo-dark.svg',
      },
      items: [
        {
          to: '/docs/getting-started',
          label: 'Install',
          position: 'right',
          activeBasePath: '/docs/getting-started',
          className: 'navbar-install-link',
        },
        {
          to: '/docs/',
          label: 'Docs',
          position: 'right',
          activeBaseRegex: '^/docs(?!/getting-started|/dms/plugins-overview).*',
        },
        {
          to: '/plugins',
          label: 'Plugins',
          position: 'right',
          activeBasePath: '/plugins',
        },
        {
          href: 'https://wallpaper.duxlabs.dev',
          label: 'Wallpaper',
          position: 'right',
          className: 'navbar-wallpaper-link',
        },
        {
          to: '/blog',
          label: 'Updates',
          position: 'right',
        },
        {
          type: 'search',
          position: 'right',
        },
        {
          type: 'docsVersionDropdown',
          position: 'right',
          className: 'navbar-version-dropdown',
        },
        {
          href: 'https://github.com/duxlabs-dev/dux-shell',
          position: 'right',
          className: 'header-github-link',
          'aria-label': 'GitHub repository',
        },
        {
          href: 'https://discord.gg/ppWTpKmPgT',
          position: 'right',
          className: 'navbar-discord-link',
          'aria-label': 'Discord community',
        },
        {
          href: 'https://ko-fi.com/duxlabs',
          position: 'right',
          className: 'navbar-kofi-link',
          'aria-label': 'Support on Ko-fi',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          items: [
            {
              label: 'Install',
              to: '/docs/getting-started',
            },
            {
              label: 'Docs',
              to: '/docs/',
            },
            {
              label: 'Contributing',
              to: '/docs/contributing',
            },
          ],
        },
      ],
      copyright: `© ${new Date().getFullYear()} Dux Labs • MIT License`,
    },
    prism: {
      theme: dankPurpleLight,
      darkTheme: dankPurple,
      additionalLanguages: ['bash', 'json', 'yaml', 'toml', 'rust', 'python', 'javascript', 'typescript', 'qml', 'nix'],
    },
  },
};

export default config;
