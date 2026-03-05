# Server Deployment Guide

Run GoZen as a background service on macOS or Linux using a process supervisor.

## macOS — launchd

Create `~/Library/LaunchAgents/com.gozen.daemon.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.gozen.daemon</string>

  <key>ProgramArguments</key>
  <array>
    <string>/usr/local/bin/zen</string>
    <string>daemon</string>
    <string>start</string>
    <string>--foreground</string>
  </array>

  <key>RunAtLoad</key>
  <true/>

  <key>KeepAlive</key>
  <true/>

  <key>StandardOutPath</key>
  <string>/tmp/gozen-stdout.log</string>

  <key>StandardErrorPath</key>
  <string>/tmp/gozen-stderr.log</string>

  <key>EnvironmentVariables</key>
  <dict>
    <key>HOME</key>
    <string>/Users/YOUR_USERNAME</string>
  </dict>
</dict>
</plist>
```

Load and start the service:

```bash
launchctl load ~/Library/LaunchAgents/com.gozen.daemon.plist
launchctl start com.gozen.daemon
```

Stop and unload:

```bash
launchctl stop com.gozen.daemon
launchctl unload ~/Library/LaunchAgents/com.gozen.daemon.plist
```

## Linux — systemd

Create `~/.config/systemd/user/gozen.service`:

```ini
[Unit]
Description=GoZen API Proxy Daemon
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/zen daemon start --foreground
Restart=on-failure
RestartSec=5

# Graceful shutdown timeout
TimeoutStopSec=30

[Install]
WantedBy=default.target
```

Enable and start:

```bash
systemctl --user daemon-reload
systemctl --user enable gozen
systemctl --user start gozen
```

Check status and logs:

```bash
systemctl --user status gozen
journalctl --user -u gozen -f
```

Enable lingering so the service runs without an active login session:

```bash
sudo loginctl enable-linger $USER
```

## Notes

- Use `--foreground` so the supervisor manages the process lifecycle directly.
- GoZen writes a PID file to `~/.zen/zend.pid`. On restart after SIGKILL, it detects and cleans stale PID files automatically.
- Default ports: proxy on 19841, web UI on 19840. Override in `~/.zen/zen.json`.
