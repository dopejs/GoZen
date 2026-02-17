import {translate} from '@docusaurus/Translate';
import CodeBlock from '@theme/CodeBlock';
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

export function Installation() {
  return (
    <section className={styles.section}>
      <div className={styles.container}>
        <h2 className={styles.heading}>
          {translate({id: 'install.title', message: 'Quick Start'})}
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
                  {translate({id: `install.${step.key}.title`, message: stepDefaults[step.key].title})}
                </h3>
                <p className={styles.stepDesc}>
                  {translate({id: `install.${step.key}.desc`, message: stepDefaults[step.key].desc})}
                </p>
                <CodeBlock language="bash">{step.code}</CodeBlock>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
