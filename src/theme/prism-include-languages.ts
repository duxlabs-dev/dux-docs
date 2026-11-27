/**
 * Swizzled (ejected) from @docusaurus/theme-classic: same behaviour as
 * upstream, plus registration of the custom grammars (kdl, conf, pam, cron)
 * that have no prismjs component. Keep in sync with upstream when upgrading.
 */
import siteConfig from '@generated/docusaurus.config';
import type * as PrismNamespace from 'prismjs';
import {registerCustomLanguages} from './prism-custom-languages';

export default function prismIncludeLanguages(
  PrismObject: typeof PrismNamespace,
): void {
  const {
    themeConfig: {prism},
  } = siteConfig;
  const {additionalLanguages} = prism as {additionalLanguages: string[]};

  // Prism components work on the Prism instance on the window, while
  // prism-react-renderer uses its own Prism instance. We temporarily mount the
  // instance onto window, import components to enhance it, then remove it to
  // avoid polluting global namespace.
  const PrismBefore = globalThis.Prism;
  globalThis.Prism = PrismObject;

  additionalLanguages.forEach((lang) => {
    if (lang === 'php') {
      // eslint-disable-next-line global-require
      require('prismjs/components/prism-markup-templating.js');
    }
    // eslint-disable-next-line global-require, import/no-dynamic-require
    require(`prismjs/components/prism-${lang}`);
  });

  // Custom grammars last: cron embeds bash, so bash must already be loaded.
  registerCustomLanguages(PrismObject);

  // Clean up and eventually restore former globalThis.Prism object (if any)
  delete (globalThis as {Prism?: unknown}).Prism;
  if (typeof PrismBefore !== 'undefined') {
    globalThis.Prism = PrismObject;
  }
}
