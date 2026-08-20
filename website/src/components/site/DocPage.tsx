import { DOC_ORDER, SIDEBAR } from '../../sidebar';
import { translator, type SiteLocale } from '../../locales';
import type { PagePayload, SiteDocument } from '../../types';
import styles from './DocPage.module.scss';

interface Props {
  readonly locale: SiteLocale;
  readonly payload: PagePayload;
  readonly document: SiteDocument;
}

function titleFor(payload: PagePayload, slug: string, localePath: string): string {
  const entry = payload.navigation.find((item) => item.slug === slug);
  return entry?.titles[localePath] ?? entry?.titles[''] ?? slug;
}

export function DocPage({ locale, payload, document }: Props) {
  const t = translator(locale);
  const index = DOC_ORDER.indexOf(document.slug);
  const previous = index > 0 ? DOC_ORDER[index - 1] : undefined;
  const next = index >= 0 && index < DOC_ORDER.length - 1 ? DOC_ORDER[index + 1] : undefined;

  return (
    <div className={styles.layout}>
      <aside className={styles.sidebar}>
        {SIDEBAR.map((group, groupIndex) => (
          <div key={group.category ?? `group-${String(groupIndex)}`} className={styles.group}>
            {group.category !== undefined && <p className={styles.groupLabel}>{t('features.title')}</p>}
            <ul>
              {group.slugs.map((slug) => (
                <li key={slug}>
                  <a
                    href={`/docs/${slug}/`}
                    className={slug === document.slug ? styles.current : undefined}
                  >
                    {titleFor(payload, slug, locale.path)}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </aside>

      <main className={styles.main}>
        <article className={`${styles.content} doc-content`} dangerouslySetInnerHTML={{ __html: document.html }} />

        <nav className={styles.pagination}>
          {previous !== undefined ? (
            <a href={`/docs/${previous}/`}>
              <span>←</span> {titleFor(payload, previous, locale.path)}
            </a>
          ) : (
            <span />
          )}
          {next !== undefined && (
            <a href={`/docs/${next}/`} className={styles.next}>
              {titleFor(payload, next, locale.path)} <span>→</span>
            </a>
          )}
        </nav>
      </main>

      {document.tableOfContents.length > 0 && (
        <nav className={styles.toc} aria-label="On this page">
          <ul>
            {document.tableOfContents.map((item) => (
              <li key={item.id} data-level={item.level}>
                <a href={`#${item.id}`}>{item.title}</a>
              </li>
            ))}
          </ul>
        </nav>
      )}
    </div>
  );
}
