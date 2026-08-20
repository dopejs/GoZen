/** Document order, mirroring the sidebar the Docusaurus site shipped. */
export const SIDEBAR: readonly { readonly category?: string; readonly slugs: readonly string[] }[] = [
  {
    slugs: [
      'getting-started',
      'providers',
      'profiles',
      'routing',
      'bindings',
      'multi-cli',
      'web-ui',
      'config',
      'config-sync',
    ],
  },
  {
    category: 'Features',
    slugs: [
      'usage-tracking',
      'health-monitoring',
      'load-balancing',
      'webhooks',
      'compression',
      'middleware',
      'agents',
      'agent-infrastructure',
      'bot',
    ],
  },
];

export const DOC_ORDER: readonly string[] = SIDEBAR.flatMap((group) => group.slugs);
