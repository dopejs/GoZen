import { useState, type ReactNode } from 'react';

import { DocPage } from './components/site/DocPage';
import { Footer } from './components/site/Footer';
import { Header } from './components/site/Header';
import { Commands } from './components/home/Commands';
import { Features } from './components/home/Features';
import { Hero } from './components/home/Hero';
import { Installation } from './components/home/Installation';
import { writeLanguagePreference } from './language-preference';
import { localeForPath, translator } from './locales';
import type { PagePayload } from './types';

interface AppProps {
  readonly payload: PagePayload;
  readonly initialLocalePath: string;
}

export function App({ payload, initialLocalePath }: AppProps): ReactNode {
  const [localePath, setLocalePath] = useState(initialLocalePath);
  const locale = localeForPath(localePath);
  const t = translator(locale);

  const changeLocale = (path: string): void => {
    const next = localeForPath(path);
    writeLanguagePreference(next.path);
    setLocalePath(next.path);
    document.documentElement.lang = next.lang;
    document.documentElement.dir = next.dir ?? 'ltr';
  };

  // A document without a translation falls back to English rather than to a
  // blank page, so an untranslated locale still reads.
  const translated =
    payload.route.kind === 'doc'
      ? (payload.translations[localePath] ?? payload.translations[''])
      : undefined;

  return (
    <div dir={locale.dir ?? 'ltr'}>
      <Header locale={locale} onLocaleChange={changeLocale} />
      <main>
        {translated === undefined ? (
          <>
            <Hero t={t} />
            <Features t={t} />
            <Installation t={t} />
            <Commands t={t} />
          </>
        ) : (
          <DocPage locale={locale} payload={payload} document={translated} />
        )}
      </main>
      <Footer locale={locale} />
    </div>
  );
}
