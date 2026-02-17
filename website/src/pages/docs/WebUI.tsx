import { useTranslation } from "react-i18next";
import { CodeBlock } from "@/components/docs/CodeBlock";

export default function WebUI() {
  const { t } = useTranslation();

  return (
    <div>
      <h1 className="mb-4 text-3xl font-bold tracking-tight text-text-primary">
        {t("docs.webUi.title")}
      </h1>
      <p className="mb-8 text-text-secondary">{t("docs.webUi.intro")}</p>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.webUi.usageTitle")}
        </h2>
        <CodeBlock
          code={`# Open in browser (auto-starts daemon if needed)
zen web`}
          language="bash"
        />
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.webUi.featuresTitle")}
        </h2>
        <ul className="space-y-2">
          {(
            [
              "providerManage",
              "bindingManage",
              "settings",
              "syncSettings",
              "logs",
              "autocomplete",
            ] as const
          ).map((key) => (
            <li
              key={key}
              className="flex items-start gap-2 text-sm text-text-secondary"
            >
              <span className="mt-1.5 inline-block h-1.5 w-1.5 flex-shrink-0 rounded-full bg-teal" />
              {t(`docs.webUi.features.${key}`)}
            </li>
          ))}
        </ul>
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.webUi.securityTitle")}
        </h2>
        <p className="mb-4 text-sm text-text-secondary">
          {t("docs.webUi.securityDesc")}
        </p>
        <ul className="mb-6 space-y-2">
          {(
            [
              "sessionAuth",
              "bruteForce",
              "rsaEncryption",
              "localBypass",
            ] as const
          ).map((key) => (
            <li
              key={key}
              className="flex items-start gap-2 text-sm text-text-secondary"
            >
              <span className="mt-1.5 inline-block h-1.5 w-1.5 flex-shrink-0 rounded-full bg-teal" />
              {t(`docs.webUi.security.${key}`)}
            </li>
          ))}
        </ul>
        <CodeBlock
          code={`# Reset the Web UI password
zen config reset-password

# Change password via Web UI
zen web  # Settings → Change Password`}
          language="bash"
        />
      </section>
    </div>
  );
}
