import {useState} from 'react';
import Layout from '@theme/Layout';
import styles from './ReleasesPage.module.scss';

interface Release {
  tag: string;
  name: string;
  date: string;
  body: string;
  url: string;
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

function renderMarkdown(body: string): string {
  // Merge duplicate "Full Changelog" lines into one
  const changelogUrls: string[] = [];
  let cleaned = body.replace(
    /\*?\*?Full Changelog\*?\*?:\s*(https?:\/\/[^\s]+)/g,
    (_match, url) => {
      changelogUrls.push(url);
      return '';
    },
  );

  // Trim and collapse blank lines
  cleaned = cleaned.trim().replace(/\n{3,}/g, '\n\n');

  let result = cleaned
    .replace(/^### (.+)$/gm, '<h3>$1</h3>')
    .replace(/^## (.+)$/gm, '<h2>$1</h2>')
    .replace(/^# (.+)$/gm, '<h1>$1</h1>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/`(.+?)`/g, '<code>$1</code>')
    .replace(/^- (.+)$/gm, '<li>$1</li>')
    .replace(/(<li>.*<\/li>\n?)+/g, '<ul>$&</ul>')
    .replace(/(https?:\/\/[^\s<)]+)/g, '<a href="$1" target="_blank" rel="noopener noreferrer">$1</a>')
    .replace(/\n\n/g, '<br/>')
    .replace(/\n/g, ' ');

  // Remove trailing/leading <br/> tags
  result = result.replace(/^(<br\s*\/?>)+/, '').replace(/(<br\s*\/?>)+$/, '');

  // Append single merged changelog link at the end
  if (changelogUrls.length > 0) {
    const lastUrl = changelogUrls[changelogUrls.length - 1];
    result += `<br/><br/><strong>Full Changelog</strong>: <a href="${lastUrl}" target="_blank" rel="noopener noreferrer">${lastUrl}</a>`;
  }

  return result;
}

export default function ReleasesPage(props: {releases: Release[]}) {
  const releases: Release[] = props.releases || [];
  const [selectedIdx, setSelectedIdx] = useState(0);
  const selected = releases[selectedIdx];

  return (
    <Layout title="Releases" description="GoZen release notes">
      <div className={styles.page}>
        {releases.length === 0 ? (
          <p className={styles.empty}>No releases found.</p>
        ) : (
          <div className={styles.layout}>
            <aside className={styles.sidebar}>
              <h2 className={styles.sidebarTitle}>Versions</h2>
              <ul className={styles.versionList}>
                {releases.map((r, i) => (
                  <li key={r.tag}>
                    <button
                      className={`${styles.versionItem} ${i === selectedIdx ? styles.versionActive : ''}`}
                      onClick={() => setSelectedIdx(i)}
                    >
                      <span className={styles.versionTag}>{r.name}</span>
                      <span className={styles.versionDate}>{formatDate(r.date)}</span>
                    </button>
                  </li>
                ))}
              </ul>
            </aside>
            <main className={styles.content}>
              {selected && (
                <div className={styles.release}>
                  <div className={styles.releaseHeader}>
                    <a href={selected.url} target="_blank" rel="noopener noreferrer" className={styles.tag}>
                      {selected.name}
                    </a>
                    <span className={styles.date}>{formatDate(selected.date)}</span>
                  </div>
                  {selected.body && (
                    <div
                      className={styles.body}
                      dangerouslySetInnerHTML={{__html: renderMarkdown(selected.body)}}
                    />
                  )}
                </div>
              )}
            </main>
          </div>
        )}
      </div>
    </Layout>
  );
}
