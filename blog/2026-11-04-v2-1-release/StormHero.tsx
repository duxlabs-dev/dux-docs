import React, {useId, type ReactNode} from 'react';
import styles from './storm.module.css';

/* Storm's lightning: forked bolts crack in from opposite corners and frame the
   title, the way the 1.5 claw tears did. Each bolt is a tapered polygon along a
   jagged polyline, thick where it strikes from and needle-thin at the tip. */

type Pt = [number, number];

function buildBolt(pts: Pt[], halfW: number): string {
  const n = pts.length - 1;
  const width = (i: number) => {
    const t = i / n;
    return halfW * Math.pow(1 - t, 0.85) * Math.min(1, 0.35 + t * 6);
  };
  const side = (s: 1 | -1) =>
    pts.map(([x, y], i) => `${(x + s * width(i)).toFixed(1)} ${y.toFixed(1)}`);
  const right = side(1);
  const left = side(-1).reverse();
  return `M ${right[0]} ${right.slice(1).map((p) => `L ${p}`).join(' ')} ${left.map((p) => `L ${p}`).join(' ')} Z`;
}

const TRUNK: Pt[] = [[60, 0], [80, 38], [50, 76], [86, 118], [56, 150], [92, 196], [70, 228], [98, 262]];
const BRANCHES: {pts: Pt[]; w: number}[] = [
  {pts: [[50, 76], [26, 104], [38, 126], [8, 172]], w: 7},
  {pts: [[86, 118], [118, 138], [106, 164], [138, 204]], w: 6},
  {pts: [[56, 150], [34, 176], [44, 194], [26, 218]], w: 3.5},
];
const SPARKS = [
  {x: 6, y: 178, r: 1.8}, {x: 140, y: 210, r: 2.2}, {x: 124, y: 96, r: 1.5},
  {x: 100, y: 268, r: 2}, {x: 22, y: 224, r: 1.4}, {x: 30, y: 60, r: 1.2},
];

export function Lightning({className, mirror = false}: {className?: string; mirror?: boolean}) {
  const uid = useId().replace(/[^a-zA-Z0-9]/g, '');
  const glow = `storm-g-${uid}`;
  const rim = `storm-r-${uid}`;
  const recess = `storm-s-${uid}`;
  const paths = [buildBolt(TRUNK, 13), ...BRANCHES.map((b) => buildBolt(b.pts, b.w))];
  return (
    <svg viewBox="0 0 146 272" preserveAspectRatio="none" className={className} aria-hidden="true" focusable="false">
      <defs>
        <linearGradient id={glow} x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" className={styles.stopCore} stopOpacity="0" />
          <stop offset="18%" className={styles.stopCore} stopOpacity="0.98" />
          <stop offset="50%" className={styles.stopMid} stopOpacity="0.95" />
          <stop offset="80%" className={styles.stopDeep} stopOpacity="0.8" />
          <stop offset="100%" className={styles.stopDeep} stopOpacity="0" />
        </linearGradient>
        <linearGradient id={rim} x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" className={styles.stopRim} stopOpacity="0" />
          <stop offset="30%" className={styles.stopRim} stopOpacity="1" />
          <stop offset="75%" className={styles.stopMid} stopOpacity="0.85" />
          <stop offset="100%" className={styles.stopMid} stopOpacity="0" />
        </linearGradient>
        <filter id={recess} x="-70%" y="-30%" width="240%" height="160%">
          <feDropShadow dx="0" dy="0" stdDeviation="3.5" className={styles.recess} />
        </filter>
      </defs>
      <g transform={mirror ? 'translate(146,0) scale(-1,1)' : undefined} filter={`url(#${recess})`}>
        {paths.map((d, i) => (
          <g key={i}>
            <path fill={`url(#${glow})`} d={d} />
            <path fill="none" stroke={`url(#${rim})`} strokeWidth="0.8" strokeLinejoin="miter" d={d} />
          </g>
        ))}
        {SPARKS.map((p, i) => (
          <circle key={`s${i}`} cx={p.x} cy={p.y} r={p.r} fill={`url(#${rim})`} opacity="0.75" />
        ))}
      </g>
    </svg>
  );
}

function Strikes({opacity}: {opacity: number}) {
  return (
    <div className={styles.strikes} aria-hidden="true" style={{opacity}}>
      <div className={`${styles.strikeSet} ${styles.strikeLeft}`}>
        <Lightning className={styles.flicker1} />
      </div>
      <div className={`${styles.strikeSet} ${styles.strikeRight}`}>
        <Lightning className={styles.flicker2} mirror />
      </div>
    </div>
  );
}

export default function StormHero() {
  return (
    <div className={styles.hero}>
      <Strikes opacity={0.85} />
      <div className={styles.heroContent}>
        <p className={styles.heroTitle}>
          DMS <span className={styles.gradientText}>2.1</span>
        </p>
        <p className={styles.heroSubtitle}>Operation Storm</p>
      </div>
    </div>
  );
}

export function StormThankYou({children}: {children: ReactNode}) {
  return (
    <div className={`${styles.hero} ${styles.thankYou}`}>
      <Strikes opacity={0.6} />
      <div className={styles.heroContent}>{children}</div>
    </div>
  );
}

export const stormStyles = styles;
