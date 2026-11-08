import React, { useState, useEffect } from 'react';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import Head from '@docusaurus/Head';
import { Check, ChevronRight } from 'material-symbols/rounded';
import LandingBackground from '../components/LandingBackground';
import ThemePreview from '../components/ThemePreview';
import SponsorStrip from '../components/SponsorStrip';
import ContributorStrip from '../components/ContributorStrip';
import styles from './index.module.css';

declare global {
  interface Window {
    mediumZoom?: (target: string | HTMLElement | NodeListOf<HTMLElement>, options?: any) => any;
    playerjs?: any;
  }
}

const compositors = [
  { name: 'niri', logo: '/img/niri.svg', duration: 600 },
  { name: 'Hyprland', logo: '/img/hyprland.svg', duration: 600 },
  { name: 'mangowm', logo: '/img/mango.png', duration: 600 },
  { name: 'Sway', logo: '/img/sway.svg', duration: 600 },
  { name: 'labwc', logo: '/img/labwc.png', duration: 600 },
  { name: 'Miracle', logo: '/img/miraclewm.svg', duration: 600 },
  { name: 'Wayland', logo: null, duration: 0 },
];

const compositorLinks: Record<string, string> = {
  'niri': 'https://github.com/niri-wm/niri',
  'Hyprland': 'https://hyprland.org/',
  'mangowm': 'https://github.com/DreamMaoMao/mangowc',
  'Sway': 'https://swaywm.org/',
  'labwc': 'https://labwc.github.io/',
  'Miracle': 'https://github.com/miracle-wm-org/miracle-wm',
};

export default function Home() {
  const [typed, setTyped] = useState('');
  const [showCursor, setShowCursor] = useState(true);
  const [reduceMotion, setReduceMotion] = useState(false);
  const [currentCompositor, setCurrentCompositor] = useState(-1);
  const [copied, setCopied] = useState(false);
  const fullText = 'curl -fsSL https://install.duxlabs.dev | sh';
  const videoRef = React.useRef<HTMLIFrameElement>(null);

  const handleCopyCommand = async () => {
    try {
      await navigator.clipboard.writeText(fullText);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy text: ', err);
    }
  };

  useEffect(() => {
    const preference = window.matchMedia('(prefers-reduced-motion: reduce)');
    const updateMotion = () => setReduceMotion(preference.matches);
    updateMotion();
    preference.addEventListener('change', updateMotion);
    return () => preference.removeEventListener('change', updateMotion);
  }, []);

  useEffect(() => {
    if (reduceMotion) {
      setTyped(fullText);
      return;
    }
    if (typed.length < fullText.length) {
      const timeout = setTimeout(() => {
        setTyped(fullText.slice(0, typed.length + 1));
      }, 50);
      return () => clearTimeout(timeout);
    }
  }, [typed, fullText, reduceMotion]);

  useEffect(() => {
    const interval = setInterval(() => {
      setShowCursor((prev) => !prev);
    }, 500);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    if (typeof window === 'undefined') return;

    const initZoom = async () => {
      try {
        await new Promise(resolve => setTimeout(resolve, 800));

        let mediumZoom: any;

        if (window.mediumZoom) {
          mediumZoom = window.mediumZoom;
        } else {
          try {
            const zoomModule = await import('medium-zoom');
            mediumZoom = zoomModule.default || zoomModule;
          } catch (e) {
            return;
          }
        }

        if (!mediumZoom) return;

        const isLight = document.documentElement.getAttribute('data-theme') === 'light';
        const background = isLight
          ? 'rgba(248, 247, 251, 0.95)'
          : 'rgba(17, 17, 17, 0.95)';

        let zoomableImages = document.querySelectorAll('img[data-zoom]');

        if (zoomableImages.length < 10) {
          await new Promise(resolve => setTimeout(resolve, 500));
          zoomableImages = document.querySelectorAll('img[data-zoom]');
        }

        if (zoomableImages.length > 0) {
          mediumZoom(zoomableImages, {
            background,
            margin: 24,
          });
        }
      } catch (err) {
        console.warn('Could not initialize image zoom:', err);
      }
    };

    initZoom();

    const observer = new MutationObserver(() => {
      initZoom();
    });

    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['data-theme']
    });

    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    if (reduceMotion) {
      setCurrentCompositor(compositors.length - 1);
      return;
    }
    const timeouts: NodeJS.Timeout[] = [];

    const showFirstTimeout = setTimeout(() => {
      setCurrentCompositor(0);
    }, 800);
    timeouts.push(showFirstTimeout);

    let cumulativeDelay = 800;
    compositors.forEach((compositor, index) => {
      if (index < compositors.length - 1) {
        cumulativeDelay += compositor.duration;
        const timeout = setTimeout(() => {
          setCurrentCompositor(index + 1);
        }, cumulativeDelay);
        timeouts.push(timeout);
      }
    });

    return () => {
      timeouts.forEach(t => clearTimeout(t));
    };
  }, [reduceMotion]);

  useEffect(() => {
    if (!videoRef.current || typeof window === 'undefined') return;

    let player: any = null;

    const script = document.createElement('script');
    script.src = '//assets.mediadelivery.net/playerjs/playerjs-latest.min.js';
    script.async = true;

    script.onload = () => {
      // @ts-ignore - playerjs is loaded from external script
      if (window.playerjs && videoRef.current) {
        // @ts-ignore
        player = new window.playerjs.Player(videoRef.current);

        player.on('ready', () => {
          const observer = new IntersectionObserver(
            (entries) => {
              entries.forEach((entry) => {
                if (entry.isIntersecting && player) {
                  player.play();
                } else if (player) {
                  player.pause();
                }
              });
            },
            { threshold: 0.5 }
          );

          if (videoRef.current) {
            observer.observe(videoRef.current);
          }
        });
      }
    };

    document.head.appendChild(script);

    return () => {
      if (script.parentNode) {
        script.parentNode.removeChild(script);
      }
    };
  }, []);

  // Scroll reveal: cards below the fold start small and scale in as they enter view.
  useEffect(() => {
    if (typeof window === 'undefined' || !('IntersectionObserver' in window)) return;

    const selector = `.${styles.screenshotCard}, .${styles.featureCard}, .${styles.supportCard}`;
    const viewportHeight = window.innerHeight;
    const pending = Array.from(document.querySelectorAll<HTMLElement>(selector)).filter(
      (el) => el.getBoundingClientRect().top > viewportHeight * 0.95,
    );
    pending.forEach((el) => el.setAttribute('data-reveal', 'pending'));

    const timers: number[] = [];
    const observer = new IntersectionObserver(
      (entries) => {
        let order = 0;
        entries
          .filter((entry) => entry.isIntersecting)
          .forEach((entry) => {
            const el = entry.target as HTMLElement;
            observer.unobserve(el);
            timers.push(window.setTimeout(() => el.setAttribute('data-reveal', 'in'), order++ * 70));
          });
      },
      { threshold: 0.12, rootMargin: '0px 0px -6% 0px' },
    );
    pending.forEach((el) => observer.observe(el));

    return () => {
      observer.disconnect();
      timers.forEach((t) => window.clearTimeout(t));
      pending.forEach((el) => el.removeAttribute('data-reveal'));
    };
  }, []);

  return (
    <Layout
      wrapperClassName="landing-main-wrapper"
      title="Modern Desktop Suite"
      description="A modern Linux desktop suite with beautiful widgets and powerful monitoring - optimized for niri, Hyprland, mangowm, Sway, and Miracle WM.">
      <Head>
        <meta property="og:image" content="https://duxlabs.dev/img/homepage/dux-preview.png" />
        <meta property="og:image:width" content="1200" />
        <meta property="og:image:height" content="630" />
        <meta name="twitter:image" content="https://duxlabs.dev/img/homepage/dux-preview.png" />
      </Head>
      <noscript>
        <style>{`
          .noscript-warning {
            background: #fef3c7;
            border-left: 4px solid #f59e0b;
            color: #92400e;
            padding: 1rem 1.5rem;
            margin: 1rem;
            border-radius: 0.5rem;
            text-align: center;
            font-size: 0.95rem;
          }
          [data-theme='dark'] .noscript-warning {
            background: #422006;
            border-left-color: #f59e0b;
            color: #fef3c7;
          }
        `}</style>
        <div className="noscript-warning">
          ⚠️ JavaScript is disabled. Some interactive features and animations on this site require JavaScript to function properly.
        </div>
      </noscript>
      <div className={styles.container}>
        <LandingBackground />
        <div className={styles.content}>
          <section className={styles.hero}>
            <div className={styles.heroContent}>
              <h1 className={styles.heroTitle}>
                <span className={styles.heroLine}>Modern Desktop</span>
                <span className={styles.heroLine}>for</span>
                <span className={styles.compositorRotatorWrapper}>
                  <noscript>
                    <span style={{display: 'inline', fontSize: 'inherit', color: 'inherit'}}>
                      niri, Hyprland, Sway, and Wayland
                    </span>
                  </noscript>
                  <span className={styles.compositorRotator}>
                    {compositors.map((compositor, index) => (
                      <span
                        key={compositor.name}
                        className={`${styles.compositorSlide} ${
                          index === currentCompositor ? styles.compositorActive : ''
                        }`}
                      >
                        {compositor.logo && (
                          <img
                            src={compositor.logo}
                            alt={compositor.name}
                            className={styles.compositorLogo}
                          />
                        )}
                        <span className={styles.compositorName}>{compositor.name}</span>
                      </span>
                    ))}
                  </span>
                </span>
              </h1>

              <div className={styles.heroCTA}>
                <Link to="/docs/getting-started" className={styles.primaryCTA}>
                  <span>Get Started</span>
                  <ChevronRight size={20} className={styles.ctaArrow} />
                </Link>
                <Link to="/docs/" className={styles.secondaryCTA}>
                  Documentation
                </Link>
              </div>

              <div className={styles.terminalFloat}>
                <div
                  className={styles.terminalWindow}
                  onClick={handleCopyCommand}
                  role="button"
                  tabIndex={0}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') handleCopyCommand();
                  }}
                  aria-label={`Copy install command: ${fullText}`}
                >
                  {copied && (
                    <div className={styles.copiedIndicator}>
                      <Check size={16} style={{ marginRight: '0.5rem' }} />
                      Copied
                    </div>
                  )}
                  <div className={styles.terminalHeader}>
                    <div className={styles.terminalLogos}>
                      {compositors
                        .filter((compositor) => compositor.logo)
                        .map((compositor) => (
                          <a
                            key={compositor.name}
                            href={compositorLinks[compositor.name]}
                            target="_blank"
                            rel="noopener noreferrer"
                            title={compositor.name}
                            className={styles.terminalLogoLink}
                            onClick={(e) => e.stopPropagation()}
                          >
                            <img
                              src={compositor.logo!}
                              alt={compositor.name}
                              className={styles.terminalLogo}
                            />
                          </a>
                        ))}
                    </div>
                  </div>
                  <div className={styles.terminalBody}>
                    <div className={styles.terminalLine}>
                      <span className={styles.prompt}>❯</span>
                      <span className={styles.typedCommand}>
                        {typed.length > 0 && (
                          <>
                            <span className={styles.cmdCommand}>{typed.slice(0, Math.min(4, typed.length))}</span>
                            {typed.length > 4 && (
                              <span className={styles.cmdFlag}>{typed.slice(4, Math.min(10, typed.length))}</span>
                            )}
                            {typed.length > 10 && (
                              <span className={styles.cmdUrl}>{typed.slice(10, Math.min(38, typed.length))}</span>
                            )}
                            {typed.length > 38 && (
                              <span className={styles.cmdPipe}>{typed.slice(38, Math.min(41, typed.length))}</span>
                            )}
                            {typed.length > 41 && (
                              <span className={styles.cmdCommand}>{typed.slice(41)}</span>
                            )}
                          </>
                        )}
                      </span>
                      <span className={`${styles.terminalCursor} ${!showCursor ? styles.hidden : ''}`}>█</span>
                    </div>
                    <div className={`${styles.terminalLine} ${typed.length >= fullText.length ? styles.fadeIn : styles.hidden}`}>
                      <span className={styles.output}>→ Detecting distribution...</span>
                    </div>
                    <div className={`${styles.terminalLine} ${typed.length >= fullText.length ? styles.fadeIn : styles.hidden}`} style={{ animationDelay: '0.3s' }}>
                      <span className={styles.success}>✓ Installing dependencies</span>
                    </div>
                    <div className={`${styles.terminalLine} ${typed.length >= fullText.length ? styles.fadeIn : styles.hidden}`} style={{ animationDelay: '0.6s' }}>
                      <span className={styles.success}>✓ Configuring Dux Modular Shell</span>
                    </div>
                    <div className={`${styles.terminalLine} ${typed.length >= fullText.length ? styles.fadeIn : styles.hidden}`} style={{ animationDelay: '0.9s' }}>
                      <span className={styles.success}>✓ Ready to rock!</span>
                    </div>
                  </div>
                </div>
              </div>

              <div className={styles.preconfiguredRow}>
                <span className={styles.preconfiguredLabel}>Pre-configured flavors</span>
                <div className={styles.preconfiguredCards}>
                  <a
                    href="https://github.com/zirconium-dev/zirconium/"
                    target="_blank"
                    rel="noopener noreferrer"
                    className={styles.preconfiguredCard}
                  >
                    <img src="/img/z.svg" alt="Zirconium" className={styles.preconfiguredLogo} />
                    <div className={styles.preconfiguredInfo}>
                      <span className={styles.preconfiguredName}>Zirconium</span>
                      <span className={styles.preconfiguredDesc}>Fedora-based immutable OS with niri + Dux Modular Shell</span>
                    </div>
                    <ChevronRight size={16} className={styles.preconfiguredArrow} />
                  </a>
                  <a
                    href="https://fedoraproject.org/spins/miraclewm/"
                    target="_blank"
                    rel="noopener noreferrer"
                    className={styles.preconfiguredCard}
                  >
                    <img src="/img/fedora-mark.svg" alt="Fedora" className={styles.preconfiguredLogo} />
                    <div className={styles.preconfiguredInfo}>
                      <span className={styles.preconfiguredName}>Fedora Miracle Spin</span>
                      <span className={styles.preconfiguredDesc}>Official Fedora Spin with MiracleWM + Dux Modular Shell</span>
                    </div>
                    <ChevronRight size={16} className={styles.preconfiguredArrow} />
                  </a>
                </div>
              </div>
            </div>
          </section>

          <section className={styles.screenshotGallery}>
            <div className={styles.sectionHeader}>
              <h2 className={styles.sectionTitle}>
                See it <span className={styles.gradientText}>in action</span>
              </h2>
              <p className={styles.sectionDesc}>
                Beautiful, functional, and ready to use
              </p>
            </div>

            <div className={styles.screenshotsGrid}>
              <div className={`${styles.screenshotCard} ${styles.large}`}>
                <div className={styles.screenshotFrame}>
                  <iframe
                    ref={videoRef}
                    className={styles.screenshotVideo}
                    src="https://player.mediadelivery.net/embed/526968/bb74c859-f2c7-43d6-aded-1715e9d3cc8e?muted=true&loop=true"
                    loading="lazy"
                    style={{ border: 0, width: '100%', height: '100%' }}
                    allow="accelerometer; gyroscope; autoplay; encrypted-media; picture-in-picture;"
                    allowFullScreen={true}
                  />
                </div>
                <div className={styles.screenshotLabel}>
                  <h3>Dux Modular Shell in Action</h3>
                  <p>Experience the fluid interface and beautiful animations</p>
                </div>
              </div>

              <div className={styles.screenshotCard}>
                <div className={styles.screenshotFrame}>
                  <img
                    src="/img/homepage/dankdash_dark.webp"
                    alt="Dux Dash - Overview Dashboard"
                    className={`${styles.screenshotImage} ${styles.darkOnly}`}
                    loading="lazy"
                    decoding="async"
                    data-zoom
                  />
                  <img
                    src="/img/homepage/dankdash_light.webp"
                    alt="Dux Dash - Overview Dashboard"
                    className={`${styles.screenshotImage} ${styles.lightOnly}`}
                    loading="lazy"
                    decoding="async"
                    data-zoom
                  />
                </div>
                <div className={styles.screenshotLabel}>
                  <h3>Dux Dash</h3>
                  <p>Media controls, weather, calendar, and system info at a glance</p>
                </div>
              </div>

              <div className={styles.screenshotCard}>
                <div className={styles.screenshotFrame}>
                  <img
                    src="/img/homepage/controlcenter_dark.webp"
                    alt="Quick Settings"
                    className={`${styles.screenshotImage} ${styles.darkOnly}`}
                    loading="lazy"
                    decoding="async"
                    data-zoom
                  />
                  <img
                    src="/img/homepage/controlcenter_light.webp"
                    alt="Quick Settings"
                    className={`${styles.screenshotImage} ${styles.lightOnly}`}
                    loading="lazy"
                    decoding="async"
                    data-zoom
                  />
                </div>
                <div className={styles.screenshotLabel}>
                  <h3>Quick Settings</h3>
                  <p>Effortlessly control your audio, displays, network, and bluetooth</p>
                </div>
              </div>

              <div className={styles.screenshotCard}>
                <div className={styles.screenshotFrame}>
                  <img
                    src="/img/homepage/widget_dark.webp"
                    alt="Widget Customization"
                    className={`${styles.screenshotImage} ${styles.darkOnly}`}
                    loading="lazy"
                    decoding="async"
                    data-zoom
                  />
                  <img
                    src="/img/homepage/widget_light.webp"
                    alt="Widget Customization"
                    className={`${styles.screenshotImage} ${styles.lightOnly}`}
                    loading="lazy"
                    decoding="async"
                    data-zoom
                  />
                </div>
                <div className={styles.screenshotLabel}>
                  <h3>Widget Customization</h3>
                  <p>Personalize your desktop experience</p>
                </div>
              </div>

              <div className={styles.screenshotCard}>
                <div className={styles.screenshotFrame}>
                  <img
                    src="/img/homepage/plugins_dark.webp"
                    alt="Plugins"
                    className={`${styles.screenshotImage} ${styles.darkOnly}`}
                    loading="lazy"
                    decoding="async"
                    data-zoom
                  />
                  <img
                    src="/img/homepage/plugins_light.webp"
                    alt="Plugins"
                    className={`${styles.screenshotImage} ${styles.lightOnly}`}
                    loading="lazy"
                    decoding="async"
                    data-zoom
                  />
                </div>
                <div className={styles.screenshotLabel}>
                  <h3>Plugins</h3>
                  <p>Extend functionality with new widgets, launcher features, and more.</p>
                </div>
              </div>
            </div>
          </section>

          <section className={styles.features}>
            <div className={styles.sectionHeader}>
              <h2 className={styles.sectionTitle}>
                Everything <span className={styles.gradientText}>you need</span>
              </h2>
              <p className={styles.sectionDesc}>
                A complete desktop experience, out of the box
              </p>
            </div>

            <div className={styles.featuresGrid}>
              <FeatureCard
                title="Dux Modular Shell"
                description="A modern and beautiful desktop shell with dynamic theming and smooth animations."
                imageDark="/img/desktop.webp"
                imageLight="/img/desktoplight.webp"
                href="/docs/dms/overview"
              />
              <FeatureCard
                title="Dux Files"
                description="A fast, keyboard-first file manager with tabs, split view, search, trash, devices, and network shares."
                imageDark="/img/dfiles.webp"
                imageLight="/img/dfileslight.webp"
                href="/docs/dfiles"
              />
              <FeatureCard
                title="Dux GOP"
                description="Stateless system and process monitoring for CPU, memory, GPU, disks, and network interfaces."
                imageDark="/img/dgop.webp"
                imageLight="/img/dgoplight.webp"
                href="/docs/dgop"
              />
              <FeatureCard
                title="Dux Greeter"
                description="An aesthetically pleasing greetd greeter for your desktop."
                imageDark="/img/dgreet.webp"
                imageLight="/img/dgreetlight.webp"
                href="/docs/dms-greeter"
              />
              <FeatureCard
                title="Dux Search"
                description="Blazingly fast and efficient file system search tool."
                imageDark="/img/dsearch.webp"
                imageLight="/img/dsearchlight.webp"
                imageAlign="top"
                href="/docs/dsearch"
              />
              <FeatureCard
                title="Dux Calendar"
                description="Your calendars, events, and tasks together, with Google, Microsoft, CalDAV, and iCloud support."
                imageDark="/img/dms/dcal_dark.webp"
                imageLight="/img/dms/dcal_light.webp"
                imageAlign="top"
                href="/docs/dcal"
              />
            </div>
          </section>

          <section className={styles.showcase}>
            <div className={styles.showcaseGrid}>
              <div className={styles.showcaseText}>
                <h2 className={styles.showcaseTitle}>
                  One theme.<br /><span className={styles.gradientText}>Across your apps.</span>
                </h2>
                <p className={styles.showcaseDesc}>
                  Dux Modular Shell automatically generates matching colors for supported GTK and Qt apps,
                  terminals, editors, and more.
                </p>
                <p className={styles.showcaseDesc}>
                  Our custom <Link to="/docs/dms/cli-dank16">dank16</Link> algorithm builds a full 16-color palette
                  for terminals and syntax highlighting, with contrast tuned for light and dark themes.
                </p>
                <ul className={styles.showcaseFeatures}>
                  <li>
                    <Check size={20} />
                    <span>GTK3, GTK4, and Qt</span>
                  </li>
                  <li>
                    <Check size={20} />
                    <span>Terminal and editor color schemes</span>
                  </li>
                  <li>
                    <Check size={20} />
                    <span>Light and dark modes</span>
                  </li>
                </ul>
                <Link to="/docs/dms/application-themes" className={styles.themingLink}>
                  Set up app theming <span aria-hidden="true">↗</span>
                </Link>
              </div>
              <ThemePreview />
            </div>
          </section>

          <section className={styles.support}>
            <SponsorStrip />
            <ContributorStrip />
            <div className={styles.supportCard}>
              <div className={styles.supportText}>
                <h2 className={styles.supportTitle}>
                  Free and open source.<br /><span className={styles.gradientText}>Kept going by you.</span>
                </h2>
                <p className={styles.supportDesc}>
                  Dux Modular Shell costs nothing to use. If it has earned a spot on your desktop,
                  a tip helps cover the time and infrastructure that keep releases coming.
                </p>
              </div>
              <div className={styles.supportActions}>
                <a
                  href="https://ko-fi.com/duxlabs"
                  target="_blank"
                  rel="noopener noreferrer"
                  className={styles.primaryCTA}
                >
                  <span className={styles.kofiIcon} aria-hidden="true" />
                  <span>Tip on Ko-fi</span>
                </a>
                <Link to="/docs/contributing" className={styles.secondaryCTA}>
                  Contribute code
                </Link>
              </div>
            </div>
          </section>
        </div>
      </div>
    </Layout>
  );
}

function FeatureCard({ title, description, imageDark, imageLight, imageAlign = 'center', href }: {
  title: string;
  description: string;
  imageDark?: string;
  imageLight?: string;
  imageAlign?: 'top' | 'center';
  href?: string;
}) {
  return (
    <div className={styles.featureCard}>
      {imageDark && imageLight && (
        <div className={styles.cardImageContainer}>
          <img
            src={imageDark}
            alt={title}
            className={`${styles.cardImage} ${styles.darkOnly} ${imageAlign === 'top' ? styles.topAlign : ''}`}
            loading="lazy"
            decoding="async"
            data-zoom
          />
          <img
            src={imageLight}
            alt={title}
            className={`${styles.cardImage} ${styles.lightOnly} ${imageAlign === 'top' ? styles.topAlign : ''}`}
            loading="lazy"
            decoding="async"
            data-zoom
          />
        </div>
      )}
      <h3 className={styles.cardTitle}>{href ? <Link to={href}>{title}<span aria-hidden="true"> ↗</span></Link> : title}</h3>
      <p className={styles.cardDesc}>{description}</p>
    </div>
  );
}
