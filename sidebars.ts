import type { SidebarsConfig } from '@docusaurus/plugin-content-docs';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */
const sidebars: SidebarsConfig = {
  tutorialSidebar: [
    'index',
    'getting-started',
    {
      type: 'doc',
      id: 'duxinstall',
      label: 'Dux Installer',
    },
    {
      type: 'category',
      label: 'Dux Shell (DMS)',
      collapsed: false,
      items: [
        {
          type: 'doc',
          id: 'dms/overview',
          label: 'Overview & Architecture',
        },
        'dms/installation',
        {
          type: 'category',
          label: 'Installation - NixOS',
          collapsed: true,
          items: [
            {
              type: 'doc',
              id: 'dms/nixos',
              label: 'NixOS module',
            },
            {
              type: 'doc',
              id: 'dms/nixos-flake',
              label: 'Flake (NixOS/home-manager)',
            },
          ],
        },
        'dms/calendar-integration',
        'dms/lock-screen-authentication',
        'dms/vpn',
        'dms/autostart',
        'dms/network-speed-monitor',
        'dms/reboot-to-another-os',
        'dms/config-profiles',
        {
          type: 'doc',
          id: 'dms/managing',
          label: 'Managing Your Installation',
        },
        'dms/compositors',
        'dms/layers',
        'dms/compositor-blur',
        {
          type: 'category',
          label: 'Themes',
          collapsed: false,
          items: [
            'dms/application-themes',
            'dms/icon-theming',
            'dms/custom-themes',
          ],
        },
        'dms/keybinds-ipc',
        {
          type: 'category',
          label: 'CLI',
          collapsed: false,
          items: [
            'dms/cli-setup',
            'dms/cli-doctor',
            {
              type: 'doc',
              id: 'dms/cli-system-updater',
              label: 'System updater',
            },
            {
              type: 'doc',
              id: 'dms/cli-switch',
              label: 'Switch channels',
            },
            {
              type: 'doc',
              id: 'dms/cli-systemd',
              label: 'Systemd service',
            },
            'dms/cli-process-management',
            'dms/cli-keybinds-cheatsheets',
            {
              type: 'doc',
              id: 'dms/cli-dank16',
              label: 'Dank16 (ANSI Colors)',
            },
            {
              type: 'doc',
              id: 'dms/cli-color-picker',
              label: 'Color Picker',
            },
            'dms/cli-brightness',
            {
              type: 'doc',
              id: 'dms/cli-icc',
              label: 'ICC Color Profiles',
            },
            {
              type: 'doc',
              id: 'dms/cli-clipboard',
              label: 'Clipboard Manager',
            },
            'dms/cli-screenshot',
            {
              type: 'doc',
              id: 'dms/cli-qr',
              label: 'QR Codes',
            },
            {
              type: 'doc',
              id: 'dms/cli-dmenu',
              label: 'Dmenu',
            },
            {
              type: 'doc',
              id: 'dms/cli-notify',
              label: 'Notifications',
            },
          ],
        },
        {
          type: 'category',
          label: 'Plugins',
          collapsed: false,
          items: [
            'dms/plugins-overview',
            'dms/plugin-development',
          ],
        },
        'dms/advanced-configuration',
      ],
    },
    {
      type: 'category',
      label: 'Dux Labs Packages',
      collapsed: false,
      items: [
        {
          type: 'doc',
          id: 'packaging/index',
          label: 'Overview',
        },
      ],
    },
    {
      type: 'category',
      label: 'Dux Greeter',
      collapsed: false,
      items: [
        {
          type: 'doc',
          id: 'dms-greeter/index',
          label: 'Overview',
        },
        'dms-greeter/installation',
        {
          type: 'category',
          label: 'Installation - NixOS',
          collapsed: true,
          items: [
            {
              type: 'doc',
              id: 'dms-greeter/nixos',
              label: 'NixOS module',
            },
            {
              type: 'doc',
              id: 'dms-greeter/nixos-flake',
              label: 'Flake',
            },
          ],
        },
        'dms-greeter/configuration',
        'dms-greeter/authentication',
      ],
    },
    {
      type: 'category',
      label: 'Dux GOP',
      collapsed: false,
      items: [
        {
          type: 'doc',
          id: 'dgop/index',
          label: 'Overview',
        },
        'dgop/installation',
        'dgop/configuration',
        'dgop/usage',
      ],
    },
    {
      type: 'category',
      label: 'Dux Search',
      collapsed: false,
      items: [
        {
          type: 'doc',
          id: 'dsearch/index',
          label: 'Overview',
        },
        'dsearch/installation',
        {
          type: 'category',
          label: 'Installation - NixOS',
          collapsed: true,
          items: [
            {
              type: 'doc',
              id: 'dsearch/nixos-flake',
              label: 'Flake (home-manager)',
            },
            {
              type: 'doc',
              id: 'dsearch/nixos',
              label: 'NixOS module',
            },
          ],
        },
        'dsearch/configuration',
        'dsearch/usage',
      ],
    },
    {
      type: 'category',
      label: 'Dux Calendar',
      collapsed: false,
      items: [
        {
          type: 'doc',
          id: 'dcal/index',
          label: 'Overview',
        },
        'dcal/installation',
        'dcal/accounts',
        'dcal/usage',
        'dcal/navigation',
        'dcal/ipc',
      ],
    },
    {
      type: 'category',
      label: 'Dux Files',
      collapsed: false,
      items: [
        {
          type: 'doc',
          id: 'dfiles/index',
          label: 'Overview',
        },
        'dfiles/installation',
        {
          type: 'category',
          label: 'Installation - NixOS',
          collapsed: true,
          items: [
            {
              type: 'doc',
              id: 'dfiles/nixos-flake',
              label: 'Flake (NixOS/home-manager)',
            },
          ],
        },
        'dfiles/usage',
        'dfiles/desktop-integration',
        'dfiles/keyboard',
        'dfiles/settings',
        'dfiles/ipc',
      ],
    },
    'contributing',
    {
      type: 'doc',
      id: 'contributing-registry',
      label: 'Contributing Plugins & Themes',
    },
    'support',
    'dms/changelog',
  ],
};

export default sidebars;
