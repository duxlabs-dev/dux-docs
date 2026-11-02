import React, {useEffect, useId, useState} from 'react';
import Link from '@docusaurus/Link';
import {ChevronRight} from 'material-symbols/rounded';
import {shapePath} from '../ExpressiveShape/paths';
import {AVATAR_SHAPES, KOFI_SUPPORTER_COUNT, fetchSponsors, type Sponsor} from '@/lib/sponsors';
import styles from './styles.module.css';

const MAX_AVATARS = 12;
const AVATAR_SIZE = 44;

// Homepage teaser for /sponsors. If the API is down only the link renders.
export default function SponsorStrip(): React.JSX.Element {
  const [sponsors, setSponsors] = useState<Sponsor[] | null>(null);
  const [total, setTotal] = useState(0);
  const clipPrefix = `sponsor-strip-${useId().replace(/[^a-zA-Z0-9_-]/g, '')}`;

  useEffect(() => {
    const controller = new AbortController();
    fetchSponsors(controller.signal)
      .then((feed) => {
        const everyone = [...feed.current, ...feed.past];
        setSponsors(everyone.slice(0, MAX_AVATARS));
        setTotal(everyone.length);
      })
      .catch(() => {
        if (!controller.signal.aborted) setSponsors([]);
      });
    return () => controller.abort();
  }, []);

  const extra = total - (sponsors?.length ?? 0);

  return (
    <section className={styles.strip} aria-labelledby="sponsor-strip-title">
      <svg className={styles.clipDefs} width="0" height="0" aria-hidden="true" focusable="false">
        <defs>
          {AVATAR_SHAPES.map((shape, index) => (
            <clipPath key={shape} id={`${clipPrefix}-${index}`} clipPathUnits="objectBoundingBox">
              <path d={shapePath(shape)} />
            </clipPath>
          ))}
        </defs>
      </svg>

      <h2 id="sponsor-strip-title" className={styles.label}>
        Backed by our sponsors and {KOFI_SUPPORTER_COUNT}+ Ko-fi supporters
      </h2>

      {sponsors === null ? (
        <div className={styles.row} aria-hidden="true">
          {Array.from({length: 8}, (_, index) => (
            <span
              key={index}
              className={`${styles.avatar} ${styles.skeleton}`}
              style={{clipPath: `url(#${clipPrefix}-${index % AVATAR_SHAPES.length})`}}
            />
          ))}
        </div>
      ) : sponsors.length > 0 ? (
        <ul className={styles.row}>
          {sponsors.map((sponsor, index) => (
            <li key={`${sponsor.source}-${sponsor.login ?? sponsor.name}-${index}`}>
              <StripAvatar sponsor={sponsor} clipPath={`url(#${clipPrefix}-${index % AVATAR_SHAPES.length})`} />
            </li>
          ))}
          {extra > 0 && (
            <li>
              <Link to="/sponsors" className={styles.more} aria-label={`${extra} more sponsors`}>
                +{extra}
              </Link>
            </li>
          )}
        </ul>
      ) : null}

      <Link to="/sponsors" className={styles.link}>
        Meet our sponsors
        <ChevronRight size={18} className={styles.arrow} />
      </Link>
    </section>
  );
}

function StripAvatar({sponsor, clipPath}: {sponsor: Sponsor; clipPath: string}) {
  const {name, login, avatarUrl, url, source} = sponsor;
  const [imageFailed, setImageFailed] = useState(false);
  const label = source === 'kofi' ? `${name} (Ko-fi supporter)` : login && login !== name ? `${name} (@${login})` : name;

  const face =
    avatarUrl && !imageFailed ? (
      <img
        src={avatarUrl}
        alt=""
        width={AVATAR_SIZE}
        height={AVATAR_SIZE}
        loading="lazy"
        decoding="async"
        onError={() => setImageFailed(true)}
      />
    ) : (
      <span className={styles.initial} aria-hidden="true">
        {(name.trim()[0] ?? '?').toUpperCase()}
      </span>
    );

  // Clip an inner span so the link's focus ring isn't clipped away.
  const shape = (
    <span className={styles.avatar} style={{clipPath}}>
      {face}
    </span>
  );

  if (url && source === 'github') {
    return (
      <a
        href={url}
        target="_blank"
        rel="noopener noreferrer"
        className={styles.member}
        title={label}
        aria-label={`${label} on GitHub, opens in a new tab`}>
        {shape}
      </a>
    );
  }
  return (
    <span className={styles.member} title={label} role="img" aria-label={label}>
      {shape}
    </span>
  );
}
