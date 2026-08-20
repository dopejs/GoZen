import type {UIKey} from '../../locales';
import {useState} from 'react';
import {Check, Copy, ArrowRight} from 'lucide-react';
import styles from './Hero.module.scss';

const installCmd =
  'curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh';

export function Hero({t}: {readonly t: (key: UIKey) => string}) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(installCmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <section className={styles.hero}>
      <div className={styles.bgGlow}>
        <div className={styles.bgGlowInner} />
      </div>
      <div className={styles.container}>
        <div className={styles.badge}>
          <span className={styles.badgeDot} />
          Open Source CLI Tool
        </div>
        <h1 className={styles.title}>
          <span className={styles.titleAccent}>GoZen</span>
          <span style={{display: 'block', marginTop: '1rem'}}>
            {t('hero.title')}
          </span>
        </h1>
        <p className={styles.subtitle}>
          {t('hero.subtitle')}
        </p>
        <p className={styles.tagline}>
          <span className={styles.taglineRow}>
            <strong className={styles.taglineLabel}>Go Zen</strong>
            {t('hero.tagline-1')}
          </span>
          <span className={styles.taglineRow}>
            <strong className={styles.taglineLabel}>Goes Env</strong>
            {t('hero.tagline-2')}
          </span>
        </p>
        <div className={styles.installBox}>
          <div className={styles.installCmd} onClick={handleCopy} role="button" tabIndex={0}>
            <span className={styles.dollar}>$</span>
            <div className={styles.cmdText}>
              <code className={styles.cmdCode}>{installCmd}</code>
            </div>
            <span className={copied ? styles.copiedIcon : styles.copyIcon}>
              {copied ? <Check size={16} /> : <Copy size={16} />}
            </span>
          </div>
        </div>
        <div className={styles.cta}>
          <a href="/docs/getting-started" className={styles.ctaPrimary}>
            {t('hero.getDocs')}
            <ArrowRight size={16} />
          </a>
          <a href="https://github.com/dopejs/gozen" target="_blank" rel="noopener noreferrer" className={styles.ctaSecondary}>
            GitHub
          </a>
        </div>
      </div>
    </section>
  );
}
