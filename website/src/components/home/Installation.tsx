import type {UIKey} from '../../locales';
import styles from './Installation.module.scss';

const steps = [
  {
    key: 'step1',
    code: 'curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh',
  },
  {key: 'step2', code: 'zen config add provider'},
  {key: 'step3', code: 'zen'},
];

const stepDefaults: Record<string, {title: string; desc: string}> = {
  step1: {title: 'Install GoZen', desc: 'One-line install script'},
  step2: {title: 'Configure Provider', desc: 'Add your first API provider'},
  step3: {title: 'Launch', desc: 'Start CLI with default configuration'},
};

export function Installation({t}: {readonly t: (key: UIKey) => string}) {
  return (
    <section className={styles.section}>
      <div className={styles.container}>
        <h2 className={styles.heading}>
          {t('install.title')}
        </h2>
        <div className={styles.steps}>
          {steps.map((step, i) => (
            <div key={step.key} className={styles.step}>
              <div className={styles.stepLeft}>
                <div className={styles.stepNumber}>{i + 1}</div>
                {i < steps.length - 1 && <div className={styles.stepLine} />}
              </div>
              <div className={styles.stepContent}>
                <h3 className={styles.stepTitle}>
                  {t(`install.${step.key}.title` as UIKey)}
                </h3>
                <p className={styles.stepDesc}>
                  {t(`install.${step.key}.desc` as UIKey)}
                </p>
                <pre><code>{step.code}</code></pre>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
