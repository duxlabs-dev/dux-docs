import React from 'react';
import Head from '@docusaurus/Head';
import useBaseUrl from '@docusaurus/useBaseUrl';
import styles from './index.module.css';

export default function DuxCalendarHome(): React.ReactNode {
  return (
    <div className={styles.page}>
      <Head>
        <title>Dux Calendar</title>
        <meta name="application-name" content="Dux Calendar" />
        <meta property="og:title" content="Dux Calendar" />
        <meta property="og:site_name" content="Dux Calendar" />
        <meta
          name="description"
          content="Dux Calendar is a free, open-source desktop calendar application for Linux and FreeBSD. It lets users connect their Google Calendar and Google Tasks accounts to view, create, update, delete, and synchronize their own calendar events and tasks."
        />
        <script type="application/ld+json">
          {JSON.stringify({
            '@context': 'https://schema.org',
            '@type': 'SoftwareApplication',
            name: 'Dux Calendar',
            operatingSystem: 'Linux, FreeBSD',
            applicationCategory: 'Productivity',
            url: 'https://duxlabs.dev/dcal/',
            author: {'@type': 'Organization', name: 'Dux Labs'},
            description:
              'Dux Calendar is a free, open-source desktop calendar application for Linux and FreeBSD. It lets users connect their Google Calendar and Google Tasks accounts to view, create, update, delete, and synchronize their own calendar events and tasks. The application runs locally, has no hosted backend, and stores calendar data and OAuth credentials on the user’s device.',
            offers: {'@type': 'Offer', price: '0', priceCurrency: 'USD'},
          })}
        </script>
      </Head>

      <main className={styles.container}>
        <header className={styles.header}>
          <img className={styles.icon} src={useBaseUrl('/img/dcal-icon.svg')} alt="Dux Calendar logo" />
          <h1 className={styles.title}>Dux Calendar</h1>
        </header>

        <p className={styles.lead}>
          Dux Calendar is a free, open-source desktop calendar application for Linux and FreeBSD. It
          lets users connect their Google Calendar and Google Tasks accounts to view, create, update,
          delete, and synchronize their own calendar events and tasks. It also supports Local,
          Microsoft, CalDAV, and iCloud calendars in the same unified agenda.
        </p>

        <p className={styles.lead}>
          Dux Calendar uses Google account data only to provide the calendar and task features
          requested by the user. The application runs locally, has no hosted backend, and stores
          calendar data and OAuth credentials on the user&apos;s device.
        </p>

        <nav className={styles.links} aria-label="Dux Calendar links">
          <a className={`${styles.linkButton} ${styles.linkPrimary}`} href="/docs/dcal/installation">
            Download Dux Calendar
          </a>
          <a className={styles.linkButton} href="/docs/dcal/">
            Documentation
          </a>
          <a className={styles.linkButton} href="/dcal/privacy">
            Privacy Policy
          </a>
          <a className={styles.linkButton} href="/dcal/terms">
            Terms of Service
          </a>
          <a className={styles.linkButton} href="https://github.com/duxlabs-dev/dux-calendar">
            Source Code
          </a>
        </nav>

        <figure className={styles.screenshot}>
          <img
            src={useBaseUrl('/img/blog/dankcalendar/preview.png')}
            alt="Dux Calendar month view with the settings and add-account dialogs open"
          />
          <figcaption>Dux Calendar&apos;s month view, settings, and account setup.</figcaption>
        </figure>

        <section className={styles.section}>
          <h2>How Dux Calendar uses Google user data</h2>
          <p>
            When a user connects a Google account, Dux Calendar accesses the user&apos;s calendar
            events and task lists through the Google Calendar and Google Tasks APIs to display and
            synchronize them in the app, and the account&apos;s email address to identify the
            connected account. All data stays on the user&apos;s device: nothing is transmitted to
            the developer or any third party. See the{' '}
            <a href="/dcal/privacy">Privacy Policy</a> for details, including how to revoke
            access.
          </p>
        </section>

        <section className={styles.section}>
          <h2>Free and open source</h2>
          <p>
            Dux Calendar is released under the MIT License and developed by Dux Labs. The
            command-line executable is named <code>dcal</code>.
          </p>
        </section>

        <footer className={styles.footer}>
          <p>
            © 2026 Dux Labs · <a href="/dcal/privacy">Privacy Policy</a> ·{' '}
            <a href="/dcal/terms">Terms of Service</a> · Contact: contact@duxlabs.dev
          </p>
        </footer>
      </main>
    </div>
  );
}
