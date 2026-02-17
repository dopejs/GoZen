import {translate} from '@docusaurus/Translate';
import {
  Terminal, Settings, Shield, GitBranch, FolderSymlink,
  Variable, Globe, RefreshCw, Sparkles, Server, Lock,
} from 'lucide-react';
import styles from './Features.module.scss';

const featureKeys = [
  {key: 'multiCli', icon: Terminal},
  {key: 'multiConfig', icon: Settings},
  {key: 'daemon', icon: Server},
  {key: 'failover', icon: Shield},
  {key: 'routing', icon: GitBranch},
  {key: 'binding', icon: FolderSymlink},
  {key: 'envVars', icon: Variable},
  {key: 'webUi', icon: Globe},
  {key: 'webSecurity', icon: Lock},
  {key: 'configSync', icon: RefreshCw},
];

const featureDefaults: Record<string, {title: string; desc: string}> = {
  multiCli: {title: 'Multi-CLI Support', desc: 'Support for Claude Code, Codex, and OpenCode with per-project configuration'},
  multiConfig: {title: 'Config Management', desc: 'Manage all API provider configs in a unified JSON file'},
  daemon: {title: 'Unified Daemon', desc: 'Single zend process hosts both the proxy server and the Web UI'},
  failover: {title: 'Proxy Failover', desc: 'Built-in HTTP proxy with automatic failover when primary provider is unavailable'},
  routing: {title: 'Scenario Routing', desc: 'Intelligent routing based on request characteristics (thinking, image, etc.)'},
  binding: {title: 'Project Bindings', desc: 'Bind directories to specific profiles and CLIs for automatic project-level configuration'},
  envVars: {title: 'Environment Variables', desc: 'Configure per-CLI environment variables at the provider level'},
  webUi: {title: 'Web Interface', desc: 'Browser-based visual management with password-protected access'},
  webSecurity: {title: 'Web Security', desc: 'Auto-generated password, session-based auth, RSA encryption for token transport'},
  configSync: {title: 'Config Sync', desc: 'Sync providers, profiles, and settings across devices with AES-256-GCM encryption'},
};

export function Features() {
  return (
    <section className={styles.section}>
      <div className={styles.container}>
        <h2 className={styles.heading}>
          {translate({id: 'features.title', message: 'Features'})}
        </h2>
        <div className={styles.grid}>
          {featureKeys.map(({key, icon: Icon}) => (
            <div key={key} className={styles.card}>
              <div className={styles.iconBox}><Icon size={20} /></div>
              <h3 className={styles.cardTitle}>
                {translate({id: `features.${key}.title`, message: featureDefaults[key].title})}
              </h3>
              <p className={styles.cardDesc}>
                {translate({id: `features.${key}.desc`, message: featureDefaults[key].desc})}
              </p>
            </div>
          ))}
          <div className={styles.comingSoon}>
            <div>
              <div className={styles.comingSoonIcon}><Sparkles size={20} /></div>
              <p className={styles.comingSoonText}>
                {translate({id: 'features.comingSoon', message: 'More features coming soon, stay tuned!'})}
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
