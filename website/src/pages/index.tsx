import Layout from '@theme/Layout';
import {translate} from '@docusaurus/Translate';
import {Hero} from '../components/home/Hero';
import {Features} from '../components/home/Features';
import {Installation} from '../components/home/Installation';
import {Commands} from '../components/home/Commands';
import styles from './index.module.scss';

export default function Home() {
  return (
    <Layout
      title={translate({id: 'homepage.title', message: 'Multi-CLI Environment Switcher'})}
      description="Manage multiple Claude Code, Codex, and OpenCode configurations with API proxy auto-failover, scenario routing, and project bindings."
    >
      <main className={styles.page}>
        <Hero />
        <Features />
        <Installation />
        <Commands />
      </main>
    </Layout>
  );
}
