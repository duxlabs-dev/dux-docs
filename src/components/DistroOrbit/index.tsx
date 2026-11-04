import React from 'react';
import ThemedImage from '@theme/ThemedImage';
import useBaseUrl from '@docusaurus/useBaseUrl';
import styles from './styles.module.css';

const distros = [
  {name: 'Arch Linux', src: '/img/distros/arch.svg'},
  {name: 'Fedora', src: '/img/distros/fedora.svg'},
  {name: 'openSUSE', src: '/img/distros/suse.svg'},
  {name: 'Debian', src: '/img/distros/debian.svg'},
  {name: 'Ubuntu', src: '/img/distros/ubuntu.svg'},
  {name: 'PikaOS', src: '/img/distros/pikaos.svg'},
  {name: 'Enterprise Linux (CentOS)', src: '/img/distros/CentOS_color_logo.svg'},
  {name: 'NixOS', src: '/img/distros/NixOS.svg'},
  {name: 'Gentoo', src: '/img/gentoo-logo.svg'},
  {name: 'Void Linux', src: '/img/distros/void.svg'},
];

export default function DistroOrbit() {
  const step = 360 / distros.length;
  return (
    <div className={styles.mashup}>
      <div className={styles.orbit}>
        <ThemedImage
          className={styles.center}
          alt="Dux Labs"
          sources={{
            light: useBaseUrl('/img/dux-labs-logo-light.svg'),
            dark: useBaseUrl('/img/dux-labs-logo-dark.svg'),
          }}
        />
        {distros.map(({name, src}, i) => (
          <div
            key={name}
            className={styles.logo}
            style={{'--angle': `${i * step}deg`} as React.CSSProperties}
            title={name}>
            <img src={useBaseUrl(src)} alt={name} />
          </div>
        ))}
      </div>
    </div>
  );
}
