# opencc

Claude Code 多环境切换器，支持 API 代理自动故障转移。

## 功能

- **多配置管理** — 在 `~/.cc_envs/` 中维护多个 API 配置，随时切换
- **代理故障转移** — 内置 HTTP 代理，当主 provider 不可用时自动切换到备用
- **TUI 配置界面** — 交互式终端界面管理配置和故障转移顺序
- **Shell 补全** — 支持 zsh / bash / fish

## 安装

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh
```

卸载：

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh -s -- --uninstall
```

## 使用

### 创建配置

在 `~/.cc_envs/` 下创建 `.env` 文件，例如 `~/.cc_envs/work.env`：

```env
ANTHROPIC_BASE_URL=https://api.anthropic.com
ANTHROPIC_AUTH_TOKEN=sk-ant-xxx
ANTHROPIC_MODEL=claude-sonnet-4-20250514
```

### 命令

```sh
# 直接使用某个配置启动 claude
opencc use <config>

# 以代理模式启动（自动读取 fallback.conf 故障转移链）
opencc

# 指定故障转移链
opencc -f provider1,provider2,provider3

# 列出所有配置
opencc list

# 打开 TUI 配置管理界面
opencc config

# 生成 shell 补全脚本
opencc completion zsh
opencc completion bash
opencc completion fish
```

### 故障转移

在 `~/.cc_envs/fallback.conf` 中配置 provider 优先级顺序（每行一个名称）：

```
work
backup
personal
```

启动时 opencc 会按顺序尝试，当前 provider 失败后自动切换到下一个。

## 开发

需要 Go 1.25+。

```sh
# 构建
go build -o opencc .

# 测试
go test ./...

# 构建当前平台二进制
./deploy.sh

# 构建所有平台二进制
./deploy.sh --all

# 然后将 dist/ 下的文件上传到 GitHub Releases
```

## 目录结构

```
├── main.go              # 入口
├── cmd/                 # CLI 命令 (cobra)
├── internal/
│   ├── config/          # fallback.conf 管理
│   ├── envfile/         # .env 文件解析
│   └── proxy/           # HTTP 代理服务器
├── tui/                 # TUI 界面 (bubbletea)
├── install.sh           # 用户安装脚本
└── deploy.sh            # 构建发布脚本
```

## License

MIT
