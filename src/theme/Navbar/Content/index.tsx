import React from 'react';
import {ThemeClassNames, useThemeConfig} from '@docusaurus/theme-common';
import {useNavbarMobileSidebar} from '@docusaurus/theme-common/internal';
import NavbarItem, {type Props as NavbarItemConfig} from '@theme/NavbarItem';
import NavbarColorModeToggle from '@theme/Navbar/ColorModeToggle';
import NavbarMobileSidebarToggle from '@theme/Navbar/MobileSidebar/Toggle';
import NavbarLogo from '@theme/Navbar/Logo';
import ExpressiveNavbarItem from '../../NavbarItem/DefaultNavbarItem';
import DocsSearch from '../../../components/DocsSearch';

export default function NavbarContent() {
  const mobileSidebar = useNavbarMobileSidebar();
  const items = useThemeConfig().navbar.items as NavbarItemConfig[];
  const destinations = items.filter(item => ('to' in item && item.to) || item.className === 'navbar-wallpaper-link');
  const utilities = items.filter(item => !('to' in item && item.to) && !item.href && item.type !== 'search');
  const socialLinks = items.filter(item => item.href && item.className !== 'navbar-wallpaper-link');

  return (
    <div className="navbar__inner">
      <div className={`${ThemeClassNames.layout.navbar.containerLeft} navbar__items`}>
        {!mobileSidebar.disabled && <NavbarMobileSidebarToggle />}
        <NavbarLogo />
      </div>
      <div className="navbar-destinations">
        {destinations.map((item, index) => <ExpressiveNavbarItem key={index} {...item} />)}
      </div>
      <div className={`${ThemeClassNames.layout.navbar.containerRight} navbar__items navbar__items--right`}>
        <div className="navbar-utilities">
          <DocsSearch />
          {utilities.map((item, index) => <NavbarItem key={index} {...item} />)}
        </div>
        <div className="navbar-actions">
          {socialLinks.map((item, index) => <NavbarItem key={index} {...item} />)}
          <NavbarColorModeToggle className="navbar-color-mode" />
        </div>
      </div>
    </div>
  );
}
