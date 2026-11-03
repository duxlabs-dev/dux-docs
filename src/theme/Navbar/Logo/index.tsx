import React from 'react';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import { useThemeConfig } from '@docusaurus/theme-common';
import ThemedImage from '@theme/ThemedImage';
import type { WrapperProps } from '@docusaurus/types';
import type LogoType from '@theme/Navbar/Logo';
import styles from './styles.module.css';

type Props = WrapperProps<typeof LogoType>;

export default function LogoWrapper(props: Props): React.JSX.Element {
  const {
    siteConfig: { title },
  } = useDocusaurusContext();
  const {
    navbar: { logo },
  } = useThemeConfig();

  const logoLink = useBaseUrl(logo?.href || '/');
  const logoImageUrl = useBaseUrl(logo?.src || '');
  const logoImageUrlDark = useBaseUrl(logo?.srcDark || logo?.src || '');

  return (
    <Link to={logoLink} className={`${styles.logoWrapper} navbar__brand`}>
      <div className="navbar__logo">
        <ThemedImage
          sources={{
            light: logoImageUrl,
            dark: logoImageUrlDark,
          }}
          alt={logo?.alt || title}
          className={styles.logo}
        />
      </div>
      <span className={`${styles.brandText} navbar-brand-text`}>dux</span>
    </Link>
  );
}
