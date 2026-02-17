import {useState} from 'react';
import {translate} from '@docusaurus/Translate';
import Link from '@docusaurus/Link';
import {Check, Copy, ArrowRight} from 'lucide-react';
import styles from './Hero.module.scss';

const installCmd =
  'curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh';

export function Hero() {
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
            {translate({id: 'hero.title', message: 'Multi-CLI Environment Switcher'})}
          </span>
        </h1>
        <p className={styles.subtitle}>
          {translate({id: 'hero.subtitle', message: 'Unified management for Claude Code, Codex, and OpenCode configurations with API proxy auto-failover'})}
        </p>
        <p className={styles.tagline}>
          <span className={styles.taglineRow}>
            <strong className={styles.taglineLabel}>Go Zen</strong>
            {translate({id: 'hero.tagline-1', message: 'enter a zen-like flow state for programming.'})}
          </span>
          <span className={styles.taglineRow}>
            <strong className={styles.taglineLabel}>Goes Env</strong>
            {translate({id: 'hero.tagline-2', message: 'seamless environment switching.'})}
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
          <Link to="/docs/getting-started" className={styles.ctaPrimary}>
            {translate({id: 'hero.getDocs', message: 'Documentation'})}
            <ArrowRight size={16} />
          </Link>
          <a href="https://github.com/dopejs/gozen" target="_blank" rel="noopener noreferrer" className={styles.ctaSecondary}>
            GitHub
          </a>
        </div>
      </div>
    </section>
  );
}
