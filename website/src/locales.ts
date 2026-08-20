import ar from './i18n/ar.json';
import de from './i18n/de.json';
import en from './i18n/en.json';
import es from './i18n/es.json';
import fr from './i18n/fr.json';
import he from './i18n/he.json';
import ja from './i18n/ja.json';
import ko from './i18n/ko.json';
import ru from './i18n/ru.json';
import zhHans from './i18n/zh-Hans.json';
import zhHant from './i18n/zh-Hant.json';

export type UIStrings = typeof en;
export type UIKey = keyof UIStrings;

/**
 * Site locales, in the shape the other dopejs sites use. `path` is the BCP 47
 * tag and the content directory name; the default locale has an empty path.
 */
export interface SiteLocale {
  readonly path: string;
  readonly lang: string;
  readonly dir?: 'rtl';
  readonly label: string;
  readonly ui: UIStrings;
}

export const SITE_LOCALES: readonly SiteLocale[] = [
  { path: '', lang: 'en', label: 'English', ui: en },
  { path: 'zh-Hans', lang: 'zh-Hans', label: '简体中文', ui: zhHans as UIStrings },
  { path: 'zh-Hant', lang: 'zh-Hant', label: '繁體中文', ui: zhHant as UIStrings },
  { path: 'es', lang: 'es', label: 'Español', ui: es as UIStrings },
  { path: 'fr', lang: 'fr', label: 'Français', ui: fr as UIStrings },
  { path: 'de', lang: 'de', label: 'Deutsch', ui: de as UIStrings },
  { path: 'ru', lang: 'ru', label: 'Русский', ui: ru as UIStrings },
  { path: 'he', lang: 'he', dir: 'rtl', label: 'עברית', ui: he as UIStrings },
  { path: 'ar', lang: 'ar', dir: 'rtl', label: 'العربية', ui: ar as UIStrings },
  { path: 'ja', lang: 'ja', label: '日本語', ui: ja as UIStrings },
  { path: 'ko', lang: 'ko', label: '한국어', ui: ko as UIStrings },
];

export function localeForPath(path: string): SiteLocale {
  return (
    SITE_LOCALES.find((locale) => locale.path === path) ??
    SITE_LOCALES.find((locale) => locale.path === '')!
  );
}

/** Translator with a fall back to English for any key a locale is missing. */
export function translator(locale: SiteLocale): (key: UIKey) => string {
  return (key) => locale.ui[key] ?? en[key];
}
