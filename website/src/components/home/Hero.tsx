import type {UIKey} from '../../locales';
import {useState} from 'react';
import {ArrowRight, Check, Copy} from 'lucide-react';
import styles from './Hero.module.scss';

const installCmd =
  'curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh';

export function Hero({t}: {readonly t: (key: UIKey) => string}) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(installCmd);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard access can be blocked by browser policy. The command remains
      // selectable in the terminal when copying is unavailable.
    }
  };

  return (
    <section className={styles.hero}>
      <div className={styles.container}>
        <div className={styles.copy}>
          <p className={styles.eyebrow}>Claude Code · Codex · OpenCode</p>
          <h1 className={styles.title}>
            <span className={styles.titleAccent}>GoZen</span>
            <span>{t('hero.title')}</span>
          </h1>
          <p className={styles.subtitle}>{t('hero.subtitle')}</p>
          <p className={styles.tagline}>
            <span><strong>Go Zen</strong>{t('hero.tagline-1')}</span>
            <span><strong>Goes Env</strong>{t('hero.tagline-2')}</span>
          </p>
          <div className={styles.cta}>
            <a href="/docs/getting-started/" className={styles.ctaPrimary}>
              {t('hero.getDocs')}
              <ArrowRight size={16} />
            </a>
            <a href="https://github.com/dopejs/GoZen" target="_blank" rel="noopener noreferrer" className={styles.ctaSecondary}>
              GitHub
            </a>
          </div>
        </div>
        <div className={styles.visual} aria-hidden="true">
          <span className={`${styles.orbit} ${styles.orbitOne}`} />
          <span className={`${styles.orbit} ${styles.orbitTwo}`} />
          <span className={styles.satelliteOne}>CLI</span>
          <span className={styles.satelliteTwo}>API</span>
          <div className={styles.logoPlate}>
            <img src="/logo.svg" alt="" width={112} height={112} />
          </div>
        </div>
      </div>

      <div className={styles.terminal}>
        <div className={styles.terminalBar}>
          <span className={styles.terminalDots} aria-hidden="true"><i /><i /><i /></span>
          <span>install.sh</span>
          <button
            type="button"
            onClick={handleCopy}
            aria-label={copied ? 'Copied' : 'Copy install command'}
          >
            {copied ? <Check size={15} /> : <Copy size={15} />}
          </button>
        </div>
        <div className={styles.terminalBody}>
          <span className={styles.prompt}>$</span>
          <code>{installCmd}</code>
        </div>
      </div>
    </section>
  );
}
