import { useTranslation } from "react-i18next";
import { CodeBlock } from "@/components/docs/CodeBlock";

const backends = [
  { key: "webdav", name: "WebDAV" },
  { key: "s3", name: "S3" },
  { key: "gist", name: "GitHub Gist" },
  { key: "repo", name: "GitHub Repo" },
];

const syncConfig = `{
  "sync": {
    "backend": "gist",
    "gist_id": "abc123def456",
    "token": "ghp_xxxxxxxxxxxx",
    "passphrase": "my-secret-passphrase",
    "auto_pull": true,
    "pull_interval": 300
  }
}`;

export default function ConfigSync() {
  const { t } = useTranslation();

  return (
    <div>
      <h1 className="mb-4 text-3xl font-bold tracking-tight text-text-primary">
        {t("docs.configSync.title")}
      </h1>
      <p className="mb-8 text-text-secondary">
        {t("docs.configSync.intro")}
      </p>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.configSync.backendsTitle")}
        </h2>
        <div className="overflow-hidden rounded-xl border border-border bg-bg-surface">
          <div className="grid grid-cols-[1fr_2fr] border-b border-border bg-bg-elevated px-5 py-3 text-xs font-semibold uppercase tracking-wider text-text-muted">
            <span>{t("docs.configSync.backendCol")}</span>
            <span>{t("docs.configSync.descCol")}</span>
          </div>
          <div className="divide-y divide-border">
            {backends.map((b) => (
              <div key={b.key} className="grid grid-cols-[1fr_2fr] px-5 py-3">
                <span className="text-sm font-medium text-text-primary">
                  {b.name}
                </span>
                <span className="text-sm text-text-secondary">
                  {t(`docs.configSync.backends.${b.key}`)}
                </span>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.configSync.setupTitle")}
        </h2>
        <p className="mb-4 text-sm text-text-secondary">
          {t("docs.configSync.setupWebUi")}
        </p>
        <CodeBlock
          code={`# Open Web UI settings
zen web  # Settings → Config Sync`}
          language="bash"
        />
        <p className="mt-4 mb-4 text-sm text-text-secondary">
          {t("docs.configSync.setupCli")}
        </p>
        <CodeBlock code="zen config sync" language="bash" />
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.configSync.configTitle")}
        </h2>
        <CodeBlock code={syncConfig} language="json" />
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.configSync.encryptionTitle")}
        </h2>
        <p className="text-sm text-text-secondary">
          {t("docs.configSync.encryptionDesc")}
        </p>
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.configSync.conflictTitle")}
        </h2>
        <ul className="space-y-2">
          {(["timestamp", "tombstone", "scalar"] as const).map((key) => (
            <li
              key={key}
              className="flex items-start gap-2 text-sm text-text-secondary"
            >
              <span className="mt-1.5 inline-block h-1.5 w-1.5 flex-shrink-0 rounded-full bg-teal" />
              {t(`docs.configSync.conflicts.${key}`)}
            </li>
          ))}
        </ul>
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.configSync.scopeTitle")}
        </h2>
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="rounded-xl border border-border bg-bg-surface p-4">
            <h3 className="mb-2 text-sm font-semibold text-sage">
              {t("docs.configSync.syncedTitle")}
            </h3>
            <ul className="space-y-1.5">
              {(["providers", "profiles", "defaultProfile", "defaultClient"] as const).map((key) => (
                <li key={key} className="flex items-start gap-2 text-sm text-text-secondary">
                  <span className="mt-1.5 inline-block h-1.5 w-1.5 flex-shrink-0 rounded-full bg-sage" />
                  {t(`docs.configSync.synced.${key}`)}
                </li>
              ))}
            </ul>
          </div>
          <div className="rounded-xl border border-border bg-bg-surface p-4">
            <h3 className="mb-2 text-sm font-semibold text-text-muted">
              {t("docs.configSync.notSyncedTitle")}
            </h3>
            <ul className="space-y-1.5">
              {(["ports", "password", "bindings", "syncConfig"] as const).map((key) => (
                <li key={key} className="flex items-start gap-2 text-sm text-text-secondary">
                  <span className="mt-1.5 inline-block h-1.5 w-1.5 flex-shrink-0 rounded-full bg-text-muted" />
                  {t(`docs.configSync.notSynced.${key}`)}
                </li>
              ))}
            </ul>
          </div>
        </div>
      </section>
    </div>
  );
}
