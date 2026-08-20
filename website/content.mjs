import { readdir, readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import matter from 'gray-matter';
import MarkdownIt from 'markdown-it';
import anchor from 'markdown-it-anchor';

const contentRoot = path.join(path.dirname(fileURLToPath(import.meta.url)), 'content');

/** Locale directory names; the default locale lives at the content root. */
export const CONTENT_LOCALES = ['zh-Hans', 'zh-Hant', 'es', 'fr', 'de', 'ru', 'he', 'ar', 'ja', 'ko'];

const markdown = new MarkdownIt({ html: true, linkify: true }).use(anchor, {
  slugify: (title) =>
    title
      .toLowerCase()
      .trim()
      .replace(/[^\p{L}\p{N}\s-]/gu, '')
      .replace(/\s+/gu, '-'),
});

function tableOfContents(tokens) {
  const items = [];
  tokens.forEach((token, index) => {
    if (token.type !== 'heading_open') return;
    const level = Number(token.tag.slice(1));
    if (level !== 2 && level !== 3) return;
    const id = token.attrGet('id');
    const title = tokens[index + 1]?.content ?? '';
    if (id !== null && title !== '') items.push({ id, level, title });
  });
  return items;
}

async function readDocument(file) {
  const source = await readFile(file, 'utf8');
  const parsed = matter(source);
  const tokens = markdown.parse(parsed.content, {});
  return {
    slug: path.basename(file, '.md'),
    title: parsed.data.title ?? path.basename(file, '.md'),
    position: parsed.data.sidebar_position ?? 99,
    html: markdown.render(parsed.content),
    tableOfContents: tableOfContents(tokens),
    text: parsed.content.replace(/[#`*_>[\]()]/gu, ' ').replace(/\s+/gu, ' ').trim(),
  };
}

async function readLocale(directory) {
  const files = (await readdir(directory)).filter((name) => name.endsWith('.md'));
  const documents = await Promise.all(files.map((name) => readDocument(path.join(directory, name))));
  return Object.fromEntries(documents.map((document) => [document.slug, document]));
}

/**
 * Every document, keyed by slug and then by locale. Routes are language-neutral,
 * so each page ships all of its translations and the client picks one.
 */
export async function loadContent() {
  const byLocale = { '': await readLocale(contentRoot) };
  for (const locale of CONTENT_LOCALES) {
    try {
      byLocale[locale] = await readLocale(path.join(contentRoot, locale));
    } catch {
      byLocale[locale] = {}; // A locale without translations falls back per document.
    }
  }
  const slugs = Object.keys(byLocale['']);
  const documents = Object.fromEntries(
    slugs.map((slug) => [
      slug,
      Object.fromEntries(
        Object.entries(byLocale)
          .map(([locale, entries]) => [locale, entries[slug]])
          .filter(([, document]) => document !== undefined),
      ),
    ]),
  );
  return { documents, slugs };
}
