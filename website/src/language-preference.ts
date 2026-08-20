import { SITE_LOCALES } from "./locales";

const STORAGE_KEY = "dopejs.locale";
const COOKIE_KEY = "dopejs_locale";

export function matchSupportedLanguage(value: string | null | undefined): string | undefined {
  if (value === null || value === undefined) return undefined;
  const normalized = value.trim().toLocaleLowerCase();
  if (["zh-hans", "zh-cn", "zh-sg"].includes(normalized)) return "zh-Hans";
  if (["zh-hant", "zh-tw", "zh-hk", "zh-mo"].includes(normalized)) return "zh-Hant";
  const exact = SITE_LOCALES.find(
    (locale) =>
      locale.path.toLocaleLowerCase() === normalized ||
      locale.lang.toLocaleLowerCase() === normalized,
  );
  if (exact !== undefined) return exact.path;
  const language = normalized.split("-")[0];
  return SITE_LOCALES.find((locale) => locale.lang.toLocaleLowerCase().split("-")[0] === language)
    ?.path;
}

function cookiePreference(): string | undefined {
  const prefix = `${COOKIE_KEY}=`;
  for (const part of document.cookie.split(";")) {
    const item = part.trim();
    if (!item.startsWith(prefix)) continue;
    return matchSupportedLanguage(decodeURIComponent(item.slice(prefix.length)));
  }
  return undefined;
}

/** Resolves explicit choice first, then the cross-subdomain cookie, then browser settings. */
export function readLanguagePreference(): string {
  try {
    const local = matchSupportedLanguage(localStorage.getItem(STORAGE_KEY));
    if (local !== undefined) return local;
  } catch {
    // Persistence can be disabled by browser policy; language detection still works.
  }
  const cookie = cookiePreference();
  if (cookie !== undefined) return cookie;
  for (const language of navigator.languages.length === 0
    ? [navigator.language]
    : navigator.languages) {
    const resolved = matchSupportedLanguage(language);
    if (resolved !== undefined) return resolved;
  }
  return "";
}

/**
 * Persists the choice for this site. The cookie is the same one the dopejs.com
 * sites share; here it stays host-only, because gozen.dev is a different
 * registrable domain and cannot read or write cookies for dopejs.com.
 */
export function writeLanguagePreference(path: string): void {
  const locale = SITE_LOCALES.find((candidate) => candidate.path === path) ?? SITE_LOCALES[0];
  const value = locale?.lang ?? "en";
  try {
    localStorage.setItem(STORAGE_KEY, value);
  } catch {
    // The domain cookie remains available when local storage is disabled.
  }
  const onDopejs = location.hostname === "dopejs.com" || location.hostname.endsWith(".dopejs.com");
  const domain = onDopejs ? "; Domain=dopejs.com" : "";
  const secure = location.protocol === "https:" ? "; Secure" : "";
  document.cookie = `${COOKIE_KEY}=${encodeURIComponent(value)}; Path=/; Max-Age=31536000; SameSite=Lax${domain}${secure}`;
}
