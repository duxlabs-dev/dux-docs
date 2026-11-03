import React, { memo } from 'react';
import clsx from 'clsx';
import { translate } from '@docusaurus/Translate';
import {
  useVisibleBlogSidebarItems,
  BlogSidebarItemList,
} from '@docusaurus/plugin-content-blog/client';
import { useLatestVersion } from '@docusaurus/plugin-content-docs/client';
import BlogSidebarContent from '@theme/BlogSidebar/Content';
import styles from '@docusaurus/theme-classic/lib/theme/BlogSidebar/Desktop/styles.module.css';

const ListComponent = ({ items }: any) => {
  return (
    <BlogSidebarItemList
      items={items}
      ulClassName={clsx(styles.sidebarItemList, 'clean-list')}
      liClassName={styles.sidebarItem}
      linkClassName={styles.sidebarItemLink}
      linkActiveClassName={styles.sidebarItemLinkActive}
    />
  );
};

function BlogSidebarDesktop({ sidebar }: any): React.JSX.Element {
  const items = useVisibleBlogSidebarItems(sidebar.items);
  let changelogUrl = '/docs/dms/changelog';

  try {
    const latestVersion = useLatestVersion('default');
    if (latestVersion) {
      const basePath = latestVersion.path === '/' ? '' : latestVersion.path;
      changelogUrl = `${basePath}/dms/changelog`;
    }
  } catch (e) {
    // fallback if docs context is unavailable on certain pages
  }

  return (
    <aside className="col col--3">
      <nav
        className={clsx(styles.sidebar, 'thin-scrollbar')}
        aria-label={translate({
          id: 'theme.blog.sidebar.navAriaLabel',
          message: 'Blog recent posts navigation',
          description: 'The ARIA label for recent posts in the blog sidebar',
        })}>
        <div style={{ paddingBottom: '1rem', borderBottom: '1px solid var(--md-sys-color-outline-variant)', marginBottom: '1.25rem' }}>
          <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--md-sys-color-on-surface-variant)', marginBottom: '0.5rem' }}>
            Latest Release Notes
          </div>
          <a
            href={changelogUrl}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              padding: '0.5rem 0.75rem',
              borderRadius: 'var(--dank-radius-lg)',
              background: 'var(--md-sys-color-primary-container)',
              color: 'var(--md-sys-color-on-primary-container)',
              fontWeight: 500,
              fontSize: '0.875rem',
              textDecoration: 'none',
            }}
          >
            <span>📋</span>
            <span>DMS changelog</span>
          </a>
        </div>

        <div className={clsx(styles.sidebarItemTitle, 'margin-bottom--md')}>
          {sidebar.title}
        </div>
        <BlogSidebarContent
          items={items}
          ListComponent={ListComponent}
          yearGroupHeadingClassName={styles.yearGroupHeading}
        />
      </nav>
    </aside>
  );
}

export default memo(BlogSidebarDesktop);
