import type {UIKey} from '../../locales';
import styles from './Features.module.scss';

const featureKeys = [
  'multiCli', 'multiConfig', 'daemon', 'failover', 'routing',
  'binding', 'envVars', 'webUi', 'webSecurity', 'configSync',
];

export function Features({t}: {readonly t: (key: UIKey) => string}) {
  return (
    <section className={styles.section}>
      <div className={styles.container}>
        <h2 className={styles.heading}>
          {t('features.title')}
        </h2>
        <div className={styles.grid}>
          {featureKeys.map((key, index) => (
            <div key={key} className={styles.card}>
              <span className={styles.number}>{String(index + 1).padStart(2, '0')}</span>
              <h3 className={styles.cardTitle}>
                {t(`features.${key}.title` as UIKey)}
              </h3>
              <p className={styles.cardDesc}>
                {t(`features.${key}.desc` as UIKey)}
              </p>
            </div>
          ))}
          <div className={styles.comingSoon}>
            <span className={styles.number}>··</span>
            <p className={styles.comingSoonText}>{t('features.comingSoon')}</p>
          </div>
        </div>
      </div>
    </section>
  );
}
