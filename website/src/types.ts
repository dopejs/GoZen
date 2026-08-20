export interface TableOfContentsItem {
  readonly id: string;
  readonly level: number;
  readonly title: string;
}

export interface SiteDocument {
  readonly slug: string;
  readonly title: string;
  readonly position: number;
  readonly html: string;
  readonly tableOfContents: readonly TableOfContentsItem[];
  readonly text: string;
}

/** One document in every locale that has it, keyed by locale path. */
export type DocumentTranslations = Readonly<Record<string, SiteDocument>>;

export interface SiteRoute {
  readonly path: string;
  readonly kind: 'home' | 'doc';
  readonly slug?: string;
}

export interface PagePayload {
  readonly route: SiteRoute;
  readonly translations: DocumentTranslations;
  /** Ordered slugs and titles per locale, for the sidebar and prev/next. */
  readonly navigation: readonly { readonly slug: string; readonly titles: Readonly<Record<string, string>> }[];
}
