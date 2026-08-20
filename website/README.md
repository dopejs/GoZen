# GoZen website

The project site for [gozen.dev](https://gozen.dev): a home page and the
documentation, in eleven languages.

Built with React 19 and Vite, on the same architecture as the other dopejs
sites. Routes are language-neutral and pre-rendered to static HTML; every
translation ships in the page payload and the visitor's language is resolved on
the client.

## Develop

Prerequisites: Node.js 22+, pnpm 10+.

```bash
pnpm install
pnpm dev        # http://localhost:5173
pnpm typecheck  # must stay at zero errors
pnpm build      # static output in dist/
pnpm preview
```

`pnpm build` runs the client build, then an SSR build, then renders each route
to static HTML with its payload embedded, and finally writes the sitemap.

## Content

Documents live in `content/`: English at the root, every other locale in
`content/<locale>/`. Order and grouping in the sidebar come from
`src/sidebar.ts`, not from the file system. Markdown is compiled by
`content.mjs` with markdown-it, and code fences are highlighted at build time by
Shiki, so no highlighter is shipped to the browser.

Interface strings live in `src/i18n/<locale>.json`, keyed the same way across
locales; English is the source and any missing key falls back to it.

To translate a document, use the helper rather than editing by hand:

```bash
python3 tools/translate-doc.py <slug> <locale> <map.json>
```

It copies code fences verbatim and fails if any prose line is missing from the
map, so a file cannot ship half-translated by accident.

> Translations are machine-generated and have not been reviewed by native
> speakers. Replacing any file under `content/<locale>/` with a reviewed
> translation needs no code change.

## Language preference

There are no per-language URLs. `src/language-preference.ts` resolves the
language from `localStorage["dopejs.locale"]`, then the `dopejs_locale` cookie,
then `navigator.languages`. The module is shared with the dopejs.com sites, but
gozen.dev is a different registrable domain, so the cookie stays host-only here
and the preference does not follow visitors between the two.

## Deploy

Pushing to `main` runs `.github/workflows/pages.yml`, which type-checks, builds
and publishes `dist/` to GitHub Pages.
