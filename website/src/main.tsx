import { createRoot, hydrateRoot } from 'react-dom/client';

import { App } from './App';
import { readLanguagePreference } from './language-preference';
import { localeForPath } from './locales';
import './css/site.scss';
import type { PagePayload } from './types';

const container = document.querySelector('#root');
if (container === null) throw new Error('GoZen site root is missing');

const embedded = document.querySelector<HTMLScriptElement>('#gozen-page');
if (embedded?.textContent == null) throw new Error('GoZen page payload is missing');
const payload = JSON.parse(embedded.textContent) as PagePayload;

const localePath = readLanguagePreference();
const locale = localeForPath(localePath);
document.documentElement.lang = locale.lang;
document.documentElement.dir = locale.dir ?? 'ltr';

const app = <App payload={payload} initialLocalePath={localePath} />;
// The static HTML is rendered in the default locale, so it can only be
// hydrated when the resolved preference is that same locale.
if (container.hasChildNodes() && localePath === '') {
  hydrateRoot(container, app);
} else {
  container.replaceChildren();
  createRoot(container).render(app);
}
