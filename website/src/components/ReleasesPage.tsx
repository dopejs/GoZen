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
  return body
    .replace(/^### (.+)$/gm, '<h3>$1</h3>')
    .replace(/^## (.+)$/gm, '<h2>$1</h2>')
    .replace(/^# (.+)$/gm, '<h1>$1</h1>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/`(.+?)`/g, '<code>$1</code>')
    .replace(/^- (.+)$/gm, '<li>$1</li>')
    .replace(/(<li>.*<\/li>\n?)+/g, '<ul>$&</ul>')
    .replace(/\n{2,}/g, '<br/><br/>')
    .replace(/\n/g, '<br/>');
}

export default function ReleasesPage(props: {releases: Release[]}) {
  const releases: Release[] = props.releases || [];

  return (
    <Layout title="Releases" description="GoZen release notes">
      <div className={styles.page}>
        <h1 className={styles.heading}>Releases</h1>
        {releases.length === 0 ? (
          <p className={styles.empty}>No releases found.</p>
        ) : (
          releases.map((r) => (
            <div key={r.tag} className={styles.release}>
              <div className={styles.releaseHeader}>
                <a href={r.url} target="_blank" rel="noopener noreferrer" className={styles.tag}>
                  {r.name}
                </a>
                <span className={styles.date}>{formatDate(r.date)}</span>
              </div>
              {r.body && (
                <div
                  className={styles.body}
                  dangerouslySetInnerHTML={{__html: renderMarkdown(r.body)}}
                />
              )}
            </div>
          ))
        )}
      </div>
    </Layout>
  );
}
