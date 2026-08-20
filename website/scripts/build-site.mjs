import { execFile } from 'node:child_process';
import { mkdir, readFile, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { promisify } from 'node:util';

import { pagePayloads } from '../content-payload.mjs';

const run = promisify(execFile);
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const output = path.join(root, 'dist');
const serverOutput = path.join(root, '.ssr');

const SITE = process.env.SITE ?? 'https://gozen.dev';

const escapeAttribute = (value) =>
  value.replaceAll('&', '&amp;').replaceAll('"', '&quot;').replaceAll('<', '&lt;');
const embeddedJson = (value) =>
  JSON.stringify(value).replaceAll('<', '\\u003c').replaceAll('-->', '--\\u003e');

await rm(output, { recursive: true, force: true });
await rm(serverOutput, { recursive: true, force: true });

try {
  await run('pnpm', ['exec', 'vite', 'build'], { cwd: root });
  await run(
    'pnpm',
    ['exec', 'vite', 'build', '--ssr', 'src/ssr.tsx', '--outDir', '.ssr', '--emptyOutDir'],
    { cwd: root },
  );

  const [{ render }, template, payloads] = await Promise.all([
    import(pathToFileURL(path.join(serverOutput, 'ssr.js')).href),
    readFile(path.join(output, 'index.html'), 'utf8'),
    pagePayloads(),
  ]);

  for (const payload of payloads) {
    const document = payload.translations[''];
    const title = document === undefined ? 'GoZen' : `${document.title} | GoZen`;
    const description =
      document === undefined
        ? 'Multi-CLI environment switcher for Claude Code, Codex, and OpenCode with API proxy auto-failover'
        : `${document.text.slice(0, 150).trim()}…`;
    const html = template
      .replace('<title>GoZen</title>', `<title>${escapeAttribute(title)}</title>`)
      .replace(
        /<meta name="description" content="[^"]*" \/?>/u,
        `<meta name="description" content="${escapeAttribute(description)}">`,
      )
      .replace('</head>', `  <link rel="canonical" href="${SITE}${payload.route.path}" />\n  </head>`)
      .replace(
        '<div id="root"></div>',
        `<div id="root">${render(payload)}</div><script id="gozen-page" type="application/json">${embeddedJson(payload)}</script>`,
      );
    const target = path.join(
      output,
      payload.route.path === '/' ? 'index.html' : `${payload.route.path.replace(/^\/|\/$/gu, '')}/index.html`,
    );
    await mkdir(path.dirname(target), { recursive: true });
    await writeFile(target, html);
  }

  const urls = payloads.map((payload) => `  <url><loc>${SITE}${payload.route.path}</loc></url>`).join('\n');
  await writeFile(
    path.join(output, 'sitemap.xml'),
    `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${urls}\n</urlset>\n`,
  );

  process.stdout.write(`site built: ${String(payloads.length)} static routes\n`);
} finally {
  await rm(serverOutput, { recursive: true, force: true });
}
