import { useEffect, useRef, useState } from 'react';
import { Github, Globe, Moon, Sun } from 'lucide-react';

import { SITE_LOCALES, translator, type SiteLocale } from '../../locales';
import styles from './Header.module.scss';

interface Props {
  readonly locale: SiteLocale;
  readonly onLocaleChange: (path: string) => void;
}

export function Header({ locale, onLocaleChange }: Props) {
  const t = translator(locale);
  const [open, setOpen] = useState(false);
  const [dark, setDark] = useState(false);
  const root = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setDark(document.documentElement.dataset.theme === 'dark');
  }, []);

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

  const toggleTheme = (): void => {
    const next = !dark;
    document.documentElement.dataset.theme = next ? 'dark' : 'light';
    document.querySelector<HTMLMetaElement>('meta[name="theme-color"]')?.setAttribute(
      'content',
      next ? '#0b1112' : '#fbfcff',
    );
    try {
      localStorage.setItem('gozen-theme', next ? 'dark' : 'light');
    } catch {
      // Theme switching still works when persistence is blocked.
    }
    setDark(next);
  };

  return (
    <header className={styles.header}>
      <div className={styles.inner}>
        <a href="/" className={styles.brand}>
          <img src="/logo.svg" alt="" aria-hidden="true" width={30} height={30} />
          <span>GoZen</span>
          <small>OSS</small>
        </a>

        <nav className={styles.nav}>
          <a className={styles.docsLink} href="/docs/getting-started/">{t('hero.getDocs')}</a>
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
          <button
            type="button"
            className={styles.iconButton}
            onClick={toggleTheme}
            aria-label={dark ? 'Use light theme' : 'Use dark theme'}
            title={dark ? 'Light theme' : 'Dark theme'}
          >
            {dark ? <Sun size={16} /> : <Moon size={16} />}
          </button>
          <a className={styles.iconButton} href="https://github.com/dopejs/GoZen" target="_blank" rel="noopener noreferrer" aria-label="GitHub">
            <Github size={16} />
          </a>
        </nav>
      </div>
    </header>
  );
}
