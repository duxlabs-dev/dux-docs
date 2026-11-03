import React from 'react';
import {useDoc} from '@docusaurus/plugin-content-docs/client';
import Composition from '../ExpressiveShape/Composition';
import styles from './styles.module.css';

interface HeroProps {
  asciiArt?: string;
  hideTitle?: boolean;
}

const projects: Record<string, string> = {
  dms: 'Dux Shell',
  'dms-greeter': 'Dux Greeter',
  dcal: 'Dux Calendar',
  dsearch: 'Dux Search',
  dgop: 'dgop',
  duxinstall: 'Dux Install',
  dankmaterialshell: 'Dux Shell',
  dankgreeter: 'Dux Greeter',
  dankcalendar: 'Dux Calendar',
  danksearch: 'Dux Search',
  dankinstall: 'Dux Install',
};

export default function Hero({hideTitle = true}: HeroProps) {
  const {metadata, contentTitle} = useDoc();
  const project = projects[metadata.id.split('/')[0]] ?? 'Dux Labs';
  const title = contentTitle ?? metadata.title;
  const showProject = !title.toLowerCase().startsWith(project.toLowerCase());
  const replacesTitle = hideTitle || Boolean(contentTitle);
  const Title = replacesTitle ? 'h1' : 'p';

  return (
    <div className={styles.hero} data-replaces-title={replacesTitle} data-content-title={Boolean(contentTitle)}>
      <div className={styles.heroContent}>
        {showProject && <p className={styles.project}>{project}</p>}
        <Title className={styles.title}>{title}</Title>
      </div>
      <Composition variant="header" className={styles.shapes} />
    </div>
  );
}
