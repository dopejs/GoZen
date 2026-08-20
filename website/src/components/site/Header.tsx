import { useEffect, useRef, useState } from 'react';
import { Github, Globe, Menu } from 'lucide-react';

import { SITE_LOCALES, translator, type SiteLocale } from '../../locales';
import styles from './Header.module.scss';

interface Props {
  readonly locale: SiteLocale;
  readonly onLocaleChange: (path: string) => void;
  readonly onToggleSidebar?: () => void;
}

export function Header({ locale, onLocaleChange, onToggleSidebar }: Props) {
  const t = translator(locale);
  const [open, setOpen] = useState(false);
  const root = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const close = (event: Event): void => {
      if (event instanceof KeyboardEvent && event.key !== 'Escape') return;
      if (event.type === 'click' && root.current?.contains(event.target as Node) === true) return;
      setOpen(false);
    };
    document.addEventListener('click', close);
    document.addEventListener('keydown', close);
    return () => {
      document.removeEventListener('click', close);
      document.removeEventListener('keydown', close);
    };
  }, [open]);

  return (
    <header className={styles.header}>
      <div className={styles.inner}>
        {onToggleSidebar !== undefined && (
          <button type="button" className={styles.menuButton} onClick={onToggleSidebar} aria-label="Menu">
            <Menu size={18} />
          </button>
        )}
        <a href="/" className={styles.brand}>
          <img src="/logo.svg" alt="" aria-hidden="true" width={26} height={26} />
          <span>GoZen</span>
        </a>

        <nav className={styles.nav}>
          <a href="/docs/getting-started/">{t('hero.getDocs')}</a>
          <a href="https://github.com/dopejs/GoZen" target="_blank" rel="noopener noreferrer" aria-label="GitHub">
            <Github size={18} />
          </a>
          <div className={styles.language} ref={root}>
            <button type="button" onClick={() => setOpen((value) => !value)} aria-expanded={open}>
              <Globe size={15} />
              <span>{locale.label}</span>
            </button>
            {open && (
              <ul>
                {SITE_LOCALES.map((candidate) => (
                  <li key={candidate.lang}>
                    <button
                      type="button"
                      lang={candidate.lang}
                      aria-current={candidate.path === locale.path ? 'true' : undefined}
                      className={candidate.path === locale.path ? styles.active : undefined}
                      onClick={() => {
                        onLocaleChange(candidate.path);
                        setOpen(false);
                      }}
                    >
                      {candidate.label}
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </nav>
      </div>
    </header>
  );
}
