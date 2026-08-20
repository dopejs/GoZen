import react from '@vitejs/plugin-react';
import { defineConfig, type Plugin } from 'vite';

import { pagePayloads } from './content-payload.mjs';

/** In dev, every route renders from index.html with its payload injected. */
function devPages(): Plugin {
  return {
    name: 'gozen-dev-pages',
    async transformIndexHtml(html, context) {
      const payloads = await pagePayloads();
      const url = context.originalUrl ?? '/';
      const path = url.split('?')[0] ?? '/';
      const payload =
        payloads.find((entry) => entry.route.path === (path.endsWith('/') ? path : `${path}/`)) ??
        payloads[0];
      return html.replace(
        '<div id="root"></div>',
        `<div id="root"></div><script id="gozen-page" type="application/json">${JSON.stringify(payload).replaceAll('<', '\\u003c')}</script>`,
      );
    },
    configureServer(server) {
      server.middlewares.use((request, _response, next) => {
        const url = request.url ?? '/';
        if (url.startsWith('/docs/') && !url.includes('.')) request.url = '/';
        next();
      });
    },
  };
}

export default defineConfig({
  plugins: [react(), devPages()],
  publicDir: 'static',
  build: { outDir: 'dist', emptyOutDir: true, sourcemap: true },
});
