import React, {useCallback, useEffect, useId, useRef, useState} from 'react';
import Layout from '@theme/Layout';
import Head from '@docusaurus/Head';
import {ChevronRight, CloudOff, Refresh} from 'material-symbols/rounded';
import {Favorite} from 'material-symbols/rounded/filled';
import LandingBackground from '../components/LandingBackground';
import ExpressiveShape from '../components/ExpressiveShape';
import {shapePath} from '../components/ExpressiveShape/paths';
import {
  AVATAR_SHAPES,
  GITHUB_SPONSORS_URL,
  KOFI_URL,
  SponsorsError,
  fetchSponsors,
  type Sponsor,
  type SponsorsFeed,
} from '@/lib/sponsors';
import styles from './sponsors.module.css';

type FeedState =
  | {status: 'loading'}
  | {status: 'error'; warmingUp: boolean}
  | {status: 'ready'; feed: SponsorsFeed};

const AVATAR_SIZE = {current: 96, past: 72} as const;
const SKELETON_COUNT = {current: 5, past: 12} as const;

export default function Sponsors(): React.JSX.Element {
  const [state, setState] = useState<FeedState>({status: 'loading'});
  const controllerRef = useRef<AbortController | null>(null);
  const clipPrefix = `sponsor-shape-${useId().replace(/[^a-zA-Z0-9_-]/g, '')}`;
  const clipFor = (index: number) => `url(#${clipPrefix}-${index % AVATAR_SHAPES.length})`;

  const load = useCallback(async () => {
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;
    setState({status: 'loading'});

    try {
      setState({status: 'ready', feed: await fetchSponsors(controller.signal)});
    } catch (err) {
      if (controller.signal.aborted) return;
      setState({status: 'error', warmingUp: err instanceof SponsorsError && err.warmingUp});
    }
  }, []);

  useEffect(() => {
    load();
    return () => controllerRef.current?.abort();
  }, [load]);

  const feed = state.status === 'ready' ? state.feed : null;
  const isEmpty = feed !== null && feed.current.length === 0 && feed.past.length === 0;

  return (
    <Layout
      title="Sponsors"
      description="The people who sponsor Dux Labs and keep Dux Shell free and open source.">
      <Head>
        <meta property="og:image" content="https://duxlabs.dev/img/homepage/dux-preview.png" />
        <meta property="og:image:width" content="1200" />
        <meta property="og:image:height" content="630" />
        <meta name="twitter:image" content="https://duxlabs.dev/img/homepage/dux-preview.png" />
      </Head>
      <div className={styles.container}>
        <LandingBackground />
        <div className={styles.content}>
          <svg className={styles.clipDefs} width="0" height="0" aria-hidden="true" focusable="false">
            <defs>
              {AVATAR_SHAPES.map((shape, index) => (
                <clipPath key={shape} id={`${clipPrefix}-${index}`} clipPathUnits="objectBoundingBox">
                  <path d={shapePath(shape)} />
                </clipPath>
              ))}
            </defs>
          </svg>

          <header className={styles.header}>
            <h1 className={styles.title}>
              Our <span className={styles.gradientText}>sponsors</span>
            </h1>
            <p className={styles.subtitle}>
              Dux is free and open source. These people help keep it that way.
            </p>
          </header>

          <p className={styles.srOnly} role="status">
            {state.status === 'loading' && 'Loading sponsors'}
            {state.status === 'error' && 'The sponsor list could not be loaded'}
            {feed && `${feed.current.length} current sponsors, ${feed.past.length} past supporters`}
          </p>

          <div aria-busy={state.status === 'loading'}>
            {state.status === 'loading' && (
              <>
                <SkeletonGroup tier="current" clipFor={clipFor} />
                <SkeletonGroup tier="past" clipFor={clipFor} />
              </>
            )}

            {state.status === 'error' && (
              <div className={styles.statusPanel}>
                <div className={styles.statusIcon} aria-hidden="true">
                  <ExpressiveShape shape="Cookie9Sided" className={styles.statusShape} />
                  <CloudOff size={28} className={styles.statusGlyph} />
                </div>
                <h2 className={styles.statusTitle}>
                  {state.warmingUp ? 'The sponsor list is warming up' : 'Couldn’t load the sponsor list'}
                </h2>
                <p className={styles.statusDesc}>
                  {state.warmingUp
                    ? 'Our server is still gathering everyone. Give it a moment and try again.'
                    : 'Something went wrong on the way here. You can still support Dux below.'}
                </p>
                <button type="button" className={styles.retryButton} onClick={load}>
                  <Refresh size={20} />
                  <span>Try again</span>
                </button>
              </div>
            )}

            {feed && feed.current.length > 0 && (
              <SponsorGroup tier="current" title="Current sponsors" sponsors={feed.current} clipFor={clipFor} />
            )}

            {feed && feed.past.length > 0 && (
              <SponsorGroup
                tier="past"
                title="Past supporters"
                sponsors={feed.past}
                clipFor={(index) => clipFor(index + 3)}
              />
            )}

            {isEmpty && (
              <div className={styles.statusPanel}>
                <h2 className={styles.statusTitle}>Be the first</h2>
                <p className={styles.statusDesc}>
                  No one is listed here yet. Your name could be the first one.
                </p>
              </div>
            )}
          </div>

          <section className={styles.become} aria-labelledby="become-a-sponsor">
            <div className={styles.becomeCard}>
              <div className={styles.becomeBlob} aria-hidden="true">
                <ExpressiveShape shape="SoftBurst" className={styles.becomeShape} />
                <Favorite size={40} className={styles.becomeGlyph} />
              </div>
              <div className={styles.becomeText}>
                <h2 id="become-a-sponsor" className={styles.becomeTitle}>
                  Become a <span className={styles.gradientText}>sponsor</span>
                </h2>
                <p className={styles.becomeDesc}>
                  If Dux makes your desktop better, consider supporting its development.
                  Every contribution helps cover the time and infrastructure behind each release.
                </p>
              </div>
              <div className={styles.becomeActions}>
                <a href={GITHUB_SPONSORS_URL} target="_blank" rel="noopener noreferrer" className={styles.primaryCTA}>
                  <span>Sponsor on GitHub</span>
                  <ChevronRight size={20} className={styles.ctaArrow} />
                </a>
                <a href={KOFI_URL} target="_blank" rel="noopener noreferrer" className={styles.secondaryCTA}>
                  <span className={styles.kofiIcon} aria-hidden="true" />
                  <span>Tip on Ko-fi</span>
                </a>
              </div>
            </div>
          </section>
        </div>
      </div>
    </Layout>
  );
}

type Tier = keyof typeof AVATAR_SIZE;

function SponsorGroup({tier, title, sponsors, clipFor}: {
  tier: Tier;
  title: string;
  sponsors: Sponsor[];
  clipFor: (index: number) => string;
}) {
  const headingId = `sponsors-${tier}`;
  return (
    <section className={`${styles.group} ${styles[tier]}`} aria-labelledby={headingId}>
      <h2 id={headingId} className={styles.groupTitle}>
        {title}
        <span className={styles.count}>{sponsors.length}</span>
      </h2>
      <ul className={styles.grid}>
        {sponsors.map((sponsor, index) => (
          <li key={`${sponsor.source}-${sponsor.login ?? sponsor.name}-${index}`} className={styles.item}>
            <SponsorTile sponsor={sponsor} size={AVATAR_SIZE[tier]} clipPath={clipFor(index)} />
          </li>
        ))}
      </ul>
    </section>
  );
}

function SponsorTile({sponsor, size, clipPath}: {sponsor: Sponsor; size: number; clipPath: string}) {
  const {name, login, avatarUrl, url, source} = sponsor;
  const isKofi = source === 'kofi';
  const handle = login && login !== name ? `@${login}` : null;
  const initial = (name.trim()[0] ?? '?').toUpperCase();
  const [imageFailed, setImageFailed] = useState(false);
  const showImage = Boolean(avatarUrl) && !imageFailed;

  const body = (
    <>
      <span className={styles.avatar}>
        <span className={`${styles.shape} ${showImage ? '' : styles.initialShape}`} style={{clipPath}}>
          {showImage ? (
            <img
              src={avatarUrl}
              alt={name}
              width={size}
              height={size}
              loading="lazy"
              decoding="async"
              onError={() => setImageFailed(true)}
            />
          ) : (
            <span className={styles.initial} aria-hidden="true">{initial}</span>
          )}
        </span>
        {isKofi && (
          <span className={styles.kofiBadge} title="Ko-fi supporter" aria-hidden="true">
            <span className={styles.kofiIcon} />
          </span>
        )}
      </span>
      <span className={styles.name} title={name}>{name}</span>
      {handle && <span className={styles.handle} title={handle}>{handle}</span>}
      {isKofi && <span className={styles.srOnly}> (Ko-fi supporter)</span>}
    </>
  );

  if (url && !isKofi) {
    return (
      <a
        href={url}
        target="_blank"
        rel="noopener noreferrer"
        className={styles.member}
        aria-label={`${name}${handle ? ` (${handle})` : ''} on GitHub, opens in a new tab`}>
        {body}
      </a>
    );
  }

  return <div className={styles.member}>{body}</div>;
}

function SkeletonGroup({tier, clipFor}: {tier: Tier; clipFor: (index: number) => string}) {
  return (
    <div className={`${styles.group} ${styles[tier]}`} aria-hidden="true">
      <div className={styles.skeletonTitle} />
      <ul className={styles.grid}>
        {Array.from({length: SKELETON_COUNT[tier]}, (_, index) => (
          <li key={index} className={styles.item}>
            <div className={styles.member}>
              <span className={styles.avatar}>
                <span
                  className={`${styles.shape} ${styles.skeleton}`}
                  style={{clipPath: clipFor(index), animationDelay: `${(index % 6) * 120}ms`}}
                />
              </span>
              <span className={styles.skeletonLabel} />
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
