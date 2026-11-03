import React, {type ComponentType, type SVGProps} from 'react';
import DefaultNavbarItem from '@theme-original/NavbarItem/DefaultNavbarItem';
import type {Props} from '@theme/NavbarItem/DefaultNavbarItem';
import {AutoStories, Campaign, DownloadForOffline, Extension, Wallpaper} from 'material-symbols/rounded';
import {
  AutoStories as AutoStoriesFilled,
  Campaign as CampaignFilled,
  DownloadForOffline as DownloadForOfflineFilled,
  Extension as ExtensionFilled,
  Wallpaper as WallpaperFilled,
} from 'material-symbols/rounded/filled';

type Icon = ComponentType<SVGProps<SVGSVGElement>>;

const destinations: Record<string, [Icon, Icon]> = {
  '/docs/getting-started': [DownloadForOffline, DownloadForOfflineFilled],
  '/docs/': [AutoStories, AutoStoriesFilled],
  '/plugins': [Extension, ExtensionFilled],
  '/blog': [Campaign, CampaignFilled],
  'https://wallpaper.duxlabs.dev': [Wallpaper, WallpaperFilled],
};

export default function ExpressiveNavbarItem(props: Props) {
  if (props.isDropdownItem || props.className?.includes('navbar-version-dropdown')) return <DefaultNavbarItem {...props} />;
  const destination = props.to || props.href;
  const icons = typeof destination === 'string' ? destinations[destination] : undefined;
  if (!icons) return <DefaultNavbarItem {...props} />;
  const [Outlined, Filled] = icons;

  return (
    <DefaultNavbarItem {...props} label={
      <>
        <span className="navbar-destination-indicator">
          <Outlined className="navbar-destination-icon navbar-icon-outlined" />
          <Filled className="navbar-destination-icon navbar-icon-filled" />
        </span>
        <span>{props.label}</span>
      </>
    } />
  );
}
