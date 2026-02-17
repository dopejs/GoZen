import { useTranslation } from "react-i18next";

const commands = [
  { cmd: "zen", key: "start" },
  { cmd: "zen -p <profile>", key: "profile" },
  { cmd: "zen -p", key: "profilePick" },
  { cmd: "zen --cli <cli>", key: "cli" },
  { cmd: "zen use <provider>", key: "use" },
  { cmd: "zen pick", key: "pick" },
  { cmd: "zen list", key: "list" },
  { cmd: "zen config add provider", key: "configAddProvider" },
  { cmd: "zen config add profile", key: "configAddProfile" },
  { cmd: "zen config default-client", key: "configDefaultClient" },
  { cmd: "zen config default-profile", key: "configDefaultProfile" },
  { cmd: "zen config reset-password", key: "configResetPassword" },
  { cmd: "zen config sync", key: "configSync" },
  { cmd: "zen daemon start", key: "daemonStart" },
  { cmd: "zen daemon stop", key: "daemonStop" },
  { cmd: "zen daemon status", key: "daemonStatus" },
  { cmd: "zen daemon enable", key: "daemonEnable" },
  { cmd: "zen daemon disable", key: "daemonDisable" },
  { cmd: "zen bind <profile>", key: "bind" },
  { cmd: "zen bind --cli <cli>", key: "bindCli" },
  { cmd: "zen unbind", key: "unbind" },
  { cmd: "zen status", key: "status" },
  { cmd: "zen web", key: "web" },
  { cmd: "zen upgrade", key: "upgrade" },
  { cmd: "zen version", key: "version" },
  { cmd: "zen completion <shell>", key: "completion" },
];

export function Commands() {
  const { t } = useTranslation();

  return (
    <section className="py-20">
      <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
        <h2 className="mb-12 text-center text-3xl font-bold tracking-tight text-text-primary">
          {t("commands.title")}
        </h2>

        <div className="no-scrollbar overflow-x-auto rounded-xl border border-border bg-bg-surface" style={{ WebkitOverflowScrolling: 'touch' }}>
          <div className="min-w-[480px]">
          <div className="grid grid-cols-[minmax(200px,1fr)_2fr] border-b border-border bg-bg-elevated px-5 py-3 text-xs font-semibold uppercase tracking-wider text-text-muted">
            <span>{t("commands.command")}</span>
            <span>{t("commands.description")}</span>
          </div>
          <div className="divide-y divide-border">
            {commands.map((item) => (
              <div
                key={item.key}
                className="grid grid-cols-[minmax(200px,1fr)_2fr] px-5 py-3 transition-colors hover:bg-bg-elevated/50"
              >
                <code className="whitespace-nowrap text-sm text-teal">{item.cmd}</code>
                <span className="text-sm text-text-secondary">
                  {t(`commands.items.${item.key}`)}
                </span>
              </div>
            ))}
          </div>
          </div>
        </div>
      </div>
    </section>
  );
}
