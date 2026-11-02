import React, {useCallback, useEffect, useId, useMemo, useRef, useState} from 'react';
import Layout from '@theme/Layout';
import Head from '@docusaurus/Head';
import Link from '@docusaurus/Link';
import {ChevronRight, CloudOff, Refresh, Translate} from 'material-symbols/rounded';
import LandingBackground from '../components/LandingBackground';
import ExpressiveShape from '../components/ExpressiveShape';
import {shapePath} from '../components/ExpressiveShape/paths';
import {AVATAR_SHAPES} from '@/lib/sponsors';
import {
  ContributorsError,
  POEDITOR_JOIN_URL,
  fetchContributors,
  languageCredits,
  sizedAvatar,
  type CodeContributor,
  type ContributorsFeed,
} from '@/lib/contributors';
import styles from './sponsors.module.css';
import own from './contributors.module.css';

type FeedState =
  | {status: 'loading'}
  | {status: 'error'; warmingUp: boolean}
  | {status: 'ready'; feed: ContributorsFeed};

// The top of the list gets full tiles; everyone after that is a compact avatar wall.
const TOP_COUNT = 24;
const AVATAR_PX = {top: 96, all: 48} as const;

export default function Contributors(): React.JSX.Element {
  const [state, setState] = useState<FeedState>({status: 'loading'});
  const controllerRef = useRef<AbortController | null>(null);
  const clipPrefix = `contributor-shape-${useId().replace(/[^a-zA-Z0-9_-]/g, '')}`;
  const clipFor = (index: number) => `url(#${clipPrefix}-${index % AVATAR_SHAPES.length})`;

  const load = useCallback(async () => {
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;
    setState({status: 'loading'});
    try {
      setState({status: 'ready', feed: await fetchContributors(controller.signal)});
    } catch (err) {
      if (controller.signal.aborted) return;
      setState({status: 'error', warmingUp: err instanceof ContributorsError && err.warmingUp});
    }
  }, []);

  useEffect(() => {
    load();
    return () => controllerRef.current?.abort();
  }, [load]);

  const feed = state.status === 'ready' ? state.feed : null;
  const languages = useMemo(() => (feed ? languageCredits(feed.translators) : []), [feed]);
  const top = feed?.code.slice(0, TOP_COUNT) ?? [];
  const rest = feed?.code.slice(TOP_COUNT) ?? [];

  return (
    <Layout
      title="Contributors"
      description="The people who write the code and translations behind Dux Shell.">
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
              Our <span className={styles.gradientText}>contributors</span>
            </h1>
            <p className={styles.subtitle}>
              Dux is built in the open. These people wrote the code and translated it into their languages.
            </p>
          </header>

          <p className={styles.srOnly} role="status">
            {state.status === 'loading' && 'Loading contributors'}
            {state.status === 'error' && 'The contributor list could not be loaded'}
            {feed && `${feed.code.length} code contributors, ${feed.translators.length} translators`}
          </p>

          <div aria-busy={state.status === 'loading'}>
            {state.status === 'loading' && (
              <div className={`${styles.group} ${styles.current}`} aria-hidden="true">
                <div className={styles.skeletonTitle} />
                <ul className={styles.grid}>
                  {Array.from({length: 10}, (_, index) => (
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
            )}

            {state.status === 'error' && (
              <div className={styles.statusPanel}>
                <div className={styles.statusIcon} aria-hidden="true">
                  <ExpressiveShape shape="Cookie9Sided" className={styles.statusShape} />
                  <CloudOff size={28} className={styles.statusGlyph} />
                </div>
                <h2 className={styles.statusTitle}>
                  {state.warmingUp ? 'The contributor list is warming up' : 'Couldn’t load the contributor list'}
                </h2>
                <p className={styles.statusDesc}>
                  {state.warmingUp
                    ? 'Our server is still gathering everyone. Give it a moment and try again.'
                    : 'Something went wrong on the way here. You can still get involved below.'}
                </p>
                <button type="button" className={styles.retryButton} onClick={load}>
                  <Refresh size={20} />
                  <span>Try again</span>
                </button>
              </div>
            )}

            {top.length > 0 && (
              <section className={`${styles.group} ${styles.current}`} aria-labelledby="contributors-top">
                <h2 id="contributors-top" className={styles.groupTitle}>
                  Code contributors
                  <span className={styles.count}>{feed?.code.length}</span>
                </h2>
                <ul className={styles.grid}>
                  {top.map((person, index) => (
                    <li key={person.login} className={styles.item}>
                      <ContributorTile person={person} px={AVATAR_PX.top} clipPath={clipFor(index)} />
                    </li>
                  ))}
                </ul>
                {rest.length > 0 && (
                  <ul className={`${styles.grid} ${own.wall}`}>
                    {rest.map((person, index) => (
                      <li key={person.login} className={own.wallItem}>
                        <ContributorTile person={person} px={AVATAR_PX.all} clipPath={clipFor(index + 3)} compact />
                      </li>
                    ))}
                  </ul>
                )}
              </section>
            )}

            {languages.length > 0 && (
              <section className={styles.group} aria-labelledby="contributors-translators">
                <h2 id="contributors-translators" className={styles.groupTitle}>
                  Translators
                  <span className={styles.count}>{feed?.translators.length}</span>
                </h2>
                <ul className={own.languages}>
                  {languages.map((lang) => (
                    <li key={lang.code} className={own.language}>
                      <h3 className={own.languageName}>
                        <span>{lang.name}</span>
                        <span className={own.languageCount}>{lang.translators.length}</span>
                      </h3>
                      <ul className={own.people}>
                        {lang.translators.map((name) => (
                          <li key={name} className={own.person}>{name}</li>
                        ))}
                      </ul>
                    </li>
                  ))}
                </ul>
              </section>
            )}
          </div>

          <section className={styles.become} aria-labelledby="become-a-contributor">
            <div className={styles.becomeCard}>
              <div className={styles.becomeBlob} aria-hidden="true">
                <ExpressiveShape shape="SoftBurst" className={styles.becomeShape} />
                <Translate size={40} className={styles.becomeGlyph} />
              </div>
              <div className={styles.becomeText}>
                <h2 id="become-a-contributor" className={styles.becomeTitle}>
                  Join <span className={styles.gradientText}>in</span>
                </h2>
                <p className={styles.becomeDesc}>
                  Fix a bug, build a feature, or bring Dux to your language. Every contribution shows up here.
                </p>
              </div>
              <div className={styles.becomeActions}>
                <Link to="/docs/contributing" className={styles.primaryCTA}>
                  <span>Start contributing</span>
                  <ChevronRight size={20} className={styles.ctaArrow} />
                </Link>
                <a href={POEDITOR_JOIN_URL} target="_blank" rel="noopener noreferrer" className={styles.secondaryCTA}>
                  <Translate size={20} />
                  <span>Translate on POEditor</span>
                </a>
              </div>
            </div>
          </section>
        </div>
      </div>
    </Layout>
  );
}

function ContributorTile({person, px, clipPath, compact = false}: {
  person: CodeContributor;
  px: number;
  clipPath: string;
  compact?: boolean;
}) {
  const {login, avatarUrl, url, contributions} = person;
  const [imageFailed, setImageFailed] = useState(false);
  const src = imageFailed ? undefined : sizedAvatar(avatarUrl, px);
  const commits = `${contributions.toLocaleString('en-US')} commit${contributions === 1 ? '' : 's'}`;

  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className={styles.member}
      title={compact ? `${login} · ${commits}` : undefined}
      aria-label={`${login}, ${commits}, on GitHub, opens in a new tab`}>
      <span className={styles.avatar}>
        <span className={`${styles.shape} ${src ? '' : styles.initialShape}`} style={{clipPath}}>
          {src ? (
            <img
              src={src}
              alt=""
              width={px}
              height={px}
              loading="lazy"
              decoding="async"
              onError={() => setImageFailed(true)}
            />
          ) : (
            <span className={styles.initial} aria-hidden="true">{(login[0] ?? '?').toUpperCase()}</span>
          )}
        </span>
      </span>
      {!compact && (
        <>
          <span className={styles.name}>{login}</span>
          <span className={styles.handle}>{commits}</span>
        </>
      )}
    </a>
  );
}
