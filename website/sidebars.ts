import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docs: [
    'getting-started',
    'providers',
    'profiles',
    'routing',
    'bindings',
    'multi-cli',
    'web-ui',
    'config',
    'config-sync',
    {
      type: 'category',
      label: 'Features',
      items: [
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
  ],
};

export default sidebars;
