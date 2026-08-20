import type {UIKey} from '../../locales';
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

export function Commands({t}: {readonly t: (key: UIKey) => string}) {
  return (
    <section className={styles.section}>
      <div className={styles.container}>
        <h2 className={styles.heading}>
          {t('commands.title')}
        </h2>
        <div className={styles.tableWrap}>
          <div className={styles.tableInner}>
            <div className={styles.tableHeader}>
              <span>{t('commands.command')}</span>
              <span>{t('commands.description')}</span>
            </div>
            <div className={styles.tableBody}>
              {commands.map((item) => (
                <div key={item.key} className={styles.tableRow}>
                  <code className={styles.cmdCode}>{item.cmd}</code>
                  <span className={styles.cmdDesc}>
                    {t(`commands.items.${item.key}` as UIKey)}
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
