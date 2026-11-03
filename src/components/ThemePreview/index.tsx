import React from 'react';
import Composition from '../ExpressiveShape/Composition';
import styles from './styles.module.css';

function WindowControls() {
  return <span className={styles.windowControls}><span>−</span><span>□</span><span>×</span></span>;
}

export default function ThemePreview() {
  return (
    <figure className={styles.preview}>
      <div className={styles.scene} role="img" aria-label="Editor and terminal illustration with matching DMS-generated syntax colors. The preview follows the site’s light or dark theme.">
        <div aria-hidden="true">
          <Composition variant="garden" className={styles.artwork} />
          <div className={styles.editor}>
            <div className={styles.titlebar}>
              <span className={styles.appIcon}>{'</>'}</span>
              <span>workspace.ts · Code</span>
              <WindowControls />
            </div>
            <div className={styles.workspace}>
              <div className={styles.explorer}>
                <span className={styles.explorerTitle}>EXPLORER</span>
                <span className={styles.folder}>⌄ studio</span>
                <span className={styles.folder}>⌄ src</span>
                <span className={styles.selectedFile}><b>TS</b> workspace.ts</span>
                <span className={styles.file}><b>TS</b> index.ts</span>
                <span className={styles.file}><b className={styles.string}>{'{}'}</b> package.json</span>
                <span className={styles.file}><b className={styles.modified}>M</b> README.md</span>
              </div>
              <div className={styles.source}>
                <div className={styles.tabs}><span><b>TS</b> workspace.ts <i>×</i></span><span>index.ts</span></div>
                <div className={styles.breadcrumb}>src <span>›</span> workspace.ts</div>
                <div className={styles.code}>
                  <div><em>1</em><span className={styles.keyword}>import</span>{' { readdir } '}<span className={styles.keyword}>from</span>{' '}<span className={styles.string}>'node:fs/promises'</span>;</div>
                  <div><em>2</em></div>
                  <div><em>3</em><span className={styles.keyword}>const</span>{' extensions = ['}<span className={styles.string}>'.qml'</span>{', '}<span className={styles.string}>'.ts'</span>{', '}<span className={styles.string}>'.go'</span>{'];'}</div>
                  <div><em>4</em></div>
                  <div><em>5</em><span className={styles.keyword}>export async function</span>{' '}<span className={styles.function}>findProjects</span>{'(path) {'}</div>
                  <div className={styles.currentLine}><em>6</em>{'  '}<span className={styles.keyword}>const</span>{' entries = '}<span className={styles.keyword}>await</span>{' '}<span className={styles.function}>readdir</span>{'(path);'}</div>
                  <div><em>7</em>{'  '}<span className={styles.keyword}>return</span>{' entries'}</div>
                  <div><em>8</em>{'    .'}<span className={styles.function}>filter</span>{'(name => !name.'}<span className={styles.function}>startsWith</span>{'('}<span className={styles.string}>'.'</span>{'))'}</div>
                  <div><em>9</em>{'    .'}<span className={styles.function}>map</span>{'(name => ({ name, path }));'}</div>
                  <div><em>10</em>{'}'}</div>
                  <div><em>11</em></div>
                  <div><em>12</em><span className={styles.keyword}>export const</span>{' formats = extensions.length;'}</div>
                </div>
              </div>
            </div>
            <div className={styles.statusbar}><span>⑂ main* <span>✓</span></span><span>UTF-8 <span>TypeScript</span></span></div>
          </div>
          <div className={styles.terminal}>
            <div className={styles.titlebar}><span className={styles.appIcon}>❯_</span><span>Terminal</span><WindowControls /></div>
            <div className={styles.terminalBody}>
              <div className={styles.prompt}><span>~/projects/studio</span><span className={styles.branch}>⑂ main</span></div>
              <div><span className={styles.string}>❯</span> npm run dev</div>
              <div className={styles.terminalOutput}><strong>VITE</strong><span>ready</span></div>
              <div><span className={styles.string}>➜</span>{'  Local:   '}<span className={styles.terminalUrl}>http://localhost:5173/</span></div>
              <div className={styles.terminalHint}>press h + enter to show help</div>
              <div className={styles.palette}>{Array.from({length: 8}, (_, index) => <span key={index} />)}</div>
            </div>
          </div>
        </div>
      </div>
      <figcaption className={styles.caption}>Matching colors, down to the syntax.</figcaption>
    </figure>
  );
}
