import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'GoZen',
  tagline: 'Multi-CLI Environment Switcher',
  favicon: 'favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://gozen.dev',
  baseUrl: '/',

  organizationName: 'dopejs',
  projectName: 'GoZen',

  onBrokenLinks: 'warn',

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en', 'zh-Hans', 'zh-Hant', 'es', 'ja', 'ko'],
    localeConfigs: {
      en: {label: 'English'},
      'zh-Hans': {label: '简体中文'},
      'zh-Hant': {label: '繁體中文'},
      es: {label: 'Español'},
      ja: {label: '日本語'},
      ko: {label: '한국어'},
    },
  },

  plugins: ['docusaurus-plugin-sass'],

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          lastVersion: '3.0',
          versions: {
            '3.0': {label: 'v3.0', path: '/'},
          },
          onlyIncludeVersions: ['3.0'],
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.scss',
        },
      } satisfies Preset.Options,
    ],
  ],

  headTags: [
    {
      tagName: 'script',
      attributes: {type: 'application/ld+json'},
      innerHTML: JSON.stringify({
        '@context': 'https://schema.org',
        '@type': 'SoftwareApplication',
        name: 'GoZen',
        alternateName: 'zen',
        description: 'Multi-CLI environment switcher for Claude Code, Codex, and OpenCode with API proxy auto-failover',
        url: 'https://gozen.dev/',
        applicationCategory: 'DeveloperApplication',
        operatingSystem: 'macOS, Linux',
        programmingLanguage: 'Go',
        license: 'https://opensource.org/licenses/MIT',
        offers: {'@type': 'Offer', price: '0', priceCurrency: 'USD'},
        codeRepository: 'https://github.com/dopejs/GoZen',
      }),
    },
  ],

  themeConfig: {
    colorMode: {
      defaultMode: 'dark',
      respectPrefersColorScheme: false,
    },
    metadata: [
      {name: 'keywords', content: 'GoZen, zen CLI, Claude Code proxy, Claude Code environment switcher, API proxy failover, multi-provider failover, Anthropic API proxy, scenario routing, project bindings'},
      {property: 'og:type', content: 'website'},
      {property: 'og:title', content: 'GoZen - Claude Code Environment Switcher & API Proxy'},
      {property: 'og:description', content: 'Manage multiple Claude Code, Codex, and OpenCode configurations with API proxy auto-failover, scenario routing, and project bindings.'},
      {property: 'og:site_name', content: 'GoZen'},
      {name: 'twitter:card', content: 'summary'},
    ],
    navbar: {
      title: 'GoZen',
      logo: {
        alt: 'GoZen Logo',
        src: 'logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docs',
          position: 'left',
          label: 'Docs',
        },
        {
          type: 'localeDropdown',
          position: 'right',
        },
        {
          href: 'https://github.com/dopejs/GoZen',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      copyright: `MIT License · Built with Go`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'json'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
