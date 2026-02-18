import {translate} from '@docusaurus/Translate';
import styles from './Commands.module.scss';

const commands = [
  {cmd: 'zen', key: 'start'},
  {cmd: 'zen -p <profile>', key: 'profile'},
  {cmd: 'zen -p', key: 'profilePick'},
  {cmd: 'zen --cli <cli>', key: 'cli'},
  {cmd: 'zen -y', key: 'yes'},
  {cmd: 'zen use <provider>', key: 'use'},
  {cmd: 'zen pick', key: 'pick'},
  {cmd: 'zen list', key: 'list'},
  {cmd: 'zen config add provider', key: 'configAddProvider'},
  {cmd: 'zen config add profile', key: 'configAddProfile'},
  {cmd: 'zen config default-client', key: 'configDefaultClient'},
  {cmd: 'zen config default-profile', key: 'configDefaultProfile'},
  {cmd: 'zen config reset-password', key: 'configResetPassword'},
  {cmd: 'zen config sync', key: 'configSync'},
  {cmd: 'zen daemon start', key: 'daemonStart'},
  {cmd: 'zen daemon stop', key: 'daemonStop'},
  {cmd: 'zen daemon status', key: 'daemonStatus'},
  {cmd: 'zen daemon enable', key: 'daemonEnable'},
  {cmd: 'zen daemon disable', key: 'daemonDisable'},
  {cmd: 'zen bind <profile>', key: 'bind'},
  {cmd: 'zen bind --cli <cli>', key: 'bindCli'},
  {cmd: 'zen unbind', key: 'unbind'},
  {cmd: 'zen status', key: 'status'},
  {cmd: 'zen web', key: 'web'},
  {cmd: 'zen upgrade', key: 'upgrade'},
  {cmd: 'zen version', key: 'version'},
  {cmd: 'zen completion <shell>', key: 'completion'},
];

const cmdDefaults: Record<string, string> = {
  start: 'Start CLI (uses project binding or default config)',
  profile: 'Start with specified profile',
  profilePick: 'Interactively select profile',
  cli: 'Use specified CLI (claude/codex/opencode)',
  yes: 'Auto-approve CLI permissions (claude --permission-mode acceptEdits, codex -a never)',
  use: 'Use specified provider directly (no proxy)',
  pick: 'Interactively select provider to start',
  list: 'List all providers and profiles',
  configAddProvider: 'Add a new provider',
  configAddProfile: 'Add a new profile',
  configDefaultClient: 'Set the default CLI client',
  configDefaultProfile: 'Set the default profile',
  configResetPassword: 'Reset the Web UI access password',
  configSync: 'Pull config from remote sync backend',
  daemonStart: 'Start the zend daemon',
  daemonStop: 'Stop the daemon',
  daemonStatus: 'Show daemon status',
  daemonEnable: 'Install daemon as system service',
  daemonDisable: 'Uninstall daemon system service',
  bind: 'Bind current directory to profile',
  bindCli: 'Bind current directory to specified CLI',
  unbind: 'Unbind current directory',
  status: 'Show current directory binding status',
  web: 'Open Web UI in browser (auto-starts daemon)',
  upgrade: 'Upgrade to latest version',
  version: 'Show version',
  completion: 'Generate shell completion (zsh/bash/fish)',
};

export function Commands() {
  return (
    <section className={styles.section}>
      <div className={styles.container}>
        <h2 className={styles.heading}>
          {translate({id: 'commands.title', message: 'Commands'})}
        </h2>
        <div className={styles.tableWrap}>
          <div className={styles.tableInner}>
            <div className={styles.tableHeader}>
              <span>{translate({id: 'commands.command', message: 'Command'})}</span>
              <span>{translate({id: 'commands.description', message: 'Description'})}</span>
            </div>
            <div className={styles.tableBody}>
              {commands.map((item) => (
                <div key={item.key} className={styles.tableRow}>
                  <code className={styles.cmdCode}>{item.cmd}</code>
                  <span className={styles.cmdDesc}>
                    {translate({id: `commands.items.${item.key}`, message: cmdDefaults[item.key]})}
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
