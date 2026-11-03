import React from 'react';
import Footer from '@theme-original/Footer';
import type FooterType from '@theme/Footer';
import type {WrapperProps} from '@docusaurus/types';

type Props = WrapperProps<typeof FooterType>;

export default function FooterWrapper(props: Props): React.JSX.Element {
  return (
    <div style={{position: 'relative'}}>
      <Footer {...props} />
      <div className="footer-icon-links">
        <a
          href="https://github.com/duxlabs-dev"
          target="_blank"
          rel="noopener noreferrer"
          className="footer-github-link"
          aria-label="GitHub repository"
        />
        <a
          href="https://discord.gg/ppWTpKmPgT"
          target="_blank"
          rel="noopener noreferrer"
          className="footer-discord-link"
          aria-label="Discord community"
        />
        <a
          href="https://ko-fi.com/duxlabs"
          target="_blank"
          rel="noopener noreferrer"
          className="footer-kofi-link"
          aria-label="Support on Ko-fi"
        />
      </div>
    </div>
  );
}
