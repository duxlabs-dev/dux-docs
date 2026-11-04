import React from 'react';
import {useDoc} from '@docusaurus/plugin-content-docs/client';
import Composition from '../ExpressiveShape/Composition';
import styles from './styles.module.css';

interface HeroProps {
  asciiArt?: string;
  hideTitle?: boolean;
  /** Replaces the default shape composition on the right of the hero. */
  visual?: React.ReactNode;
}

const projects: Record<string, string> = {
  dms: 'Dux Modular Shell',
  'dms-greeter': 'Dux Greeter',
  dcal: 'Dux Calendar',
  dsearch: 'Dux Search',
  dgop: 'dgop',
  duxinstall: 'Dux Install',
  dankmaterialshell: 'Dux Modular Shell',
  dankgreeter: 'Dux Greeter',
  dankcalendar: 'Dux Calendar',
  danksearch: 'Dux Search',
  dankinstall: 'Dux Install',
};

export default function Hero({hideTitle = true, visual}: HeroProps) {
  const {metadata, contentTitle} = useDoc();
  const project = projects[metadata.id.split('/')[0]] ?? 'Dux Labs';
  const title = contentTitle ?? metadata.title;
  const showProject = !title.toLowerCase().startsWith(project.toLowerCase());
  const replacesTitle = hideTitle || Boolean(contentTitle);
  const Title = replacesTitle ? 'h1' : 'p';

  return (
    <div className={styles.hero} data-replaces-title={replacesTitle} data-content-title={Boolean(contentTitle)} data-has-visual={Boolean(visual)}>
      <div className={styles.heroContent}>
        {showProject && <p className={styles.project}>{project}</p>}
        <Title className={styles.title}>{title}</Title>
      </div>
      {visual ? (
        <div className={styles.visual}>{visual}</div>
      ) : (
        <Composition variant="header" className={styles.shapes} />
      )}
    </div>
  );
}
