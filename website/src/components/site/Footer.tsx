import type { SiteLocale } from '../../locales';
import styles from './Footer.module.scss';

export function Footer({ locale }: { readonly locale: SiteLocale }) {
  return (
    <footer className={styles.footer} lang={locale.lang}>
      <div className={styles.inner}>
        <span>© 2026 GoZen</span>
        <a href="https://github.com/dopejs/GoZen" target="_blank" rel="noopener noreferrer">
          github.com/dopejs/GoZen
        </a>
      </div>
    </footer>
  );
}
