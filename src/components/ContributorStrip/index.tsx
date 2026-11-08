import React, {useEffect, useId, useState} from 'react';
import Link from '@docusaurus/Link';
import {ChevronRight} from 'material-symbols/rounded';
import {shapePath} from '../ExpressiveShape/paths';
import {AVATAR_SHAPES} from '@/lib/sponsors';
import {fetchContributors, sizedAvatar, type CodeContributor, type ContributorsFeed} from '@/lib/contributors';
import styles from '../SponsorStrip/styles.module.css';

const MAX_AVATARS = 12;
const AVATAR_SIZE = 44;

// Homepage teaser for /contributors, styled like SponsorStrip. If the API is down only the link renders.
export default function ContributorStrip(): React.JSX.Element {
  const [feed, setFeed] = useState<ContributorsFeed | null>(null);
  const [failed, setFailed] = useState(false);
  const clipPrefix = `contributor-strip-${useId().replace(/[^a-zA-Z0-9_-]/g, '')}`;
  const clipFor = (index: number) => `url(#${clipPrefix}-${index % AVATAR_SHAPES.length})`;

  useEffect(() => {
    const controller = new AbortController();
    fetchContributors(controller.signal)
      .then(setFeed)
      .catch(() => {
        if (!controller.signal.aborted) setFailed(true);
      });
    return () => controller.abort();
  }, []);

  const shown = feed?.code.slice(0, MAX_AVATARS) ?? [];
  const extra = (feed?.code.length ?? 0) - shown.length;

  return (
    <section className={styles.strip} aria-labelledby="contributor-strip-title">
      <svg className={styles.clipDefs} width="0" height="0" aria-hidden="true" focusable="false">
        <defs>
          {AVATAR_SHAPES.map((shape, index) => (
            <clipPath key={shape} id={`${clipPrefix}-${index}`} clipPathUnits="objectBoundingBox">
              <path d={shapePath(shape)} />
            </clipPath>
          ))}
        </defs>
      </svg>

      <h2 id="contributor-strip-title" className={styles.label}>
        {feed ? stripLabel(feed) : 'Built in the open by our contributors'}
      </h2>

      {!feed && !failed ? (
        <div className={styles.row} aria-hidden="true">
          {Array.from({length: 8}, (_, index) => (
            <span key={index} className={`${styles.avatar} ${styles.skeleton}`} style={{clipPath: clipFor(index)}} />
          ))}
        </div>
      ) : shown.length > 0 ? (
        <ul className={styles.row}>
          {shown.map((person, index) => (
            <li key={person.login}>
              <StripAvatar person={person} clipPath={clipFor(index)} />
            </li>
          ))}
          {extra > 0 && (
            <li>
              <Link to="/contributors" className={styles.more} aria-label={`${extra} more contributors`}>
                +{extra}
              </Link>
            </li>
          )}
        </ul>
      ) : null}

      <Link to="/contributors" className={styles.link}>
        Meet our contributors
        <ChevronRight size={18} className={styles.arrow} />
      </Link>
    </section>
  );
}

function stripLabel({code, translatorCount}: ContributorsFeed): string {
  const people = `${code.length} contributors`;
  return translatorCount > 0
    ? `Built by ${people} and ${translatorCount} translators`
    : `Built by ${people}`;
}

function StripAvatar({person, clipPath}: {person: CodeContributor; clipPath: string}) {
  const {login, avatarUrl, url} = person;
  const [imageFailed, setImageFailed] = useState(false);
  const src = imageFailed ? undefined : sizedAvatar(avatarUrl, AVATAR_SIZE);

  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className={styles.member}
      title={login}
      aria-label={`${login} on GitHub, opens in a new tab`}>
      {/* Clip an inner span so the link's focus ring isn't clipped away. */}
      <span className={styles.avatar} style={{clipPath}}>
        {src ? (
          <img
            src={src}
            alt=""
            width={AVATAR_SIZE}
            height={AVATAR_SIZE}
            loading="lazy"
            decoding="async"
            onError={() => setImageFailed(true)}
          />
        ) : (
          <span className={styles.initial} aria-hidden="true">
            {(login[0] ?? '?').toUpperCase()}
          </span>
        )}
      </span>
    </a>
  );
}
