import React, {type ComponentType, type SVGProps} from 'react';
import DefaultNavbarItem from '@theme-original/NavbarItem/DefaultNavbarItem';
import {useVersions} from '@docusaurus/plugin-content-docs/client';
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
  const versions = useVersions(undefined);
  if (props.isDropdownItem) {
    // Tagged here rather than in the version labels, which the dropdown button also shows.
    const version = versions.find((v) => v.label === props.label);
    const tag = version?.name === 'current' ? 'Next' : version?.isLast ? 'Stable' : '';
    if (!tag) return <DefaultNavbarItem {...props} />;
    return <DefaultNavbarItem {...props} label={<>{props.label}<span className="navbar-version-tag">{tag}</span></>} />;
  }
  if (props.className?.includes('navbar-version-dropdown')) return <DefaultNavbarItem {...props} />;
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
