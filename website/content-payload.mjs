import { loadContent } from './content.mjs';

const DOC_ORDER = [
  'getting-started', 'providers', 'profiles', 'routing', 'bindings', 'multi-cli',
  'web-ui', 'config', 'config-sync', 'usage-tracking', 'health-monitoring',
  'load-balancing', 'webhooks', 'compression', 'middleware', 'agents',
  'agent-infrastructure', 'bot',
];

/** One payload per language-neutral route: the home page and every document. */
export async function pagePayloads() {
  const { documents } = await loadContent();
  const navigation = DOC_ORDER.filter((slug) => documents[slug] !== undefined).map((slug) => ({
    slug,
    titles: Object.fromEntries(
      Object.entries(documents[slug]).map(([locale, document]) => [locale, document.title]),
    ),
  }));

  return [
    { route: { path: '/', kind: 'home' }, translations: {}, navigation },
    ...navigation.map(({ slug }) => ({
      route: { path: `/docs/${slug}/`, kind: 'doc', slug },
      translations: documents[slug],
      navigation,
    })),
  ];
}
