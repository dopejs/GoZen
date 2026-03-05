# Feature Specification: zen use Command Enhancements

**Feature Branch**: `010-use-command-enhancements`
**Created**: 2026-03-05
**Status**: Draft
**Input**: User description: "zen use xxx 这个模式无法使用 zen use xxx --yes，需要支持。

延伸思考：
1. 可以考虑将 --yes 作为一个配置，让用户设置可以默认开启yes模式。
2. 目前我们使用zen之后无法只用 cc 或者 codex 的启动参数，是否可以考虑使用 zen [zen参数] -- [client参数] 或者 zen use xx -- [client参数] 这种模式。将对应的client参数在client启动时注入进去。"

## Clarifications

### Session 2026-03-05

- Q: What permission mode should `zen --yes` use for Claude Code? → A: `--permission-mode bypassPermissions` (skips all permission prompts). Web UI configuration allows users to enable auto-permission mode with customizable settings. Priority: `--` parameters > `--yes` flag > Web UI config > default behavior.
- Q: What are Codex's permission flag options? → A: Codex uses `-a` / `--ask-for-approval` with options: `untrusted` (ask for untrusted operations), `on-request` (ask in interactive runs), `never` (never ask, for non-interactive runs). `zen --yes` should use `-a never` to match `bypassPermissions` behavior.
- Q: Should Web UI abstract permission modes across clients or show client-specific options? → A: No abstraction. Web UI displays client-specific permission options based on selected client. Claude Code shows 5 modes (default, acceptEdits, bypassPermissions, dontAsk, plan), Codex shows 3 modes (untrusted, on-request, never). Configuration stored per-client.
- Q: How to resolve conflicts when `--yes` and `--` both specify permission parameters? → A: Priority order: `--` parameters (highest) > `--yes` flag > Web UI config > default behavior. `--` explicitly expresses user intent for CLI parameters and takes precedence over all other settings.

## User Scenarios & Testing

### User Story 1 - Auto-Approve Client Permissions with --yes Flag (Priority: P1)

Users need to bypass all client permission prompts (file edits, reads, bash commands) to enable uninterrupted workflow, especially for automation scripts and power users who trust the AI assistant.

**Why this priority**: This is the core issue reported - the current `--yes` flag uses `acceptEdits` which only auto-approves file edits, but users still get interrupted by prompts for file reads and bash commands, breaking workflow continuity.

**Independent Test**: Can be fully tested by running `zen --yes` and verifying that Claude Code receives `--permission-mode bypassPermissions` (or Codex receives equivalent permission flags), resulting in zero permission prompts during the session.

**Acceptance Scenarios**:

1. **Given** a user runs `zen --yes` with Claude Code, **When** the client starts, **Then** it receives `--permission-mode bypassPermissions` and operates without any permission prompts
2. **Given** a user runs `zen --yes` with Codex, **When** the client starts, **Then** it receives the equivalent auto-approve permission flags
3. **Given** a user runs `zen --yes` in a CI/CD script, **When** the client executes tasks, **Then** no interactive prompts block the automation

---

### User Story 2 - Persistent Auto-Permission Configuration (Priority: P2)

Users who frequently work with AI assistants want to enable auto-permission mode by default through Web UI configuration, without typing `--yes` every time.

**Why this priority**: Enhances user experience for power users but doesn't block basic functionality. Users can still use `--yes` flag while this configuration option is being developed.

**Independent Test**: Can be tested independently by enabling auto-permission mode in Web UI for a specific client (e.g., Claude Code with `bypassPermissions`), and verifying that subsequent `zen` commands automatically pass the configured permission flags to that client.

**Acceptance Scenarios**:

1. **Given** a user has enabled auto-permission for Claude Code with `bypassPermissions`, **When** they run `zen` with Claude Code, **Then** the client receives `--permission-mode bypassPermissions` automatically
2. **Given** a user has enabled auto-permission for Codex with `on-request`, **When** they run `zen --cli codex`, **Then** Codex receives `-a on-request` automatically
3. **Given** a user switches between clients, **When** they run `zen --cli claude` then `zen --cli codex`, **Then** each client receives its own configured permission mode
4. **Given** a user has auto-permission enabled, **When** they run `zen --yes`, **Then** `--yes` overrides the config and uses the most permissive mode (`bypassPermissions` for Claude Code, `never` for Codex)
5. **Given** a user has not configured auto-permission for a client, **When** they run that client, **Then** it starts with default permission behavior

---

### User Story 3 - Client Parameter Pass-Through (Priority: P3)

Users want to pass client-specific startup parameters (like `cc` or `codex` flags) through the `zen` wrapper to the underlying client, with full control over all parameters including permission modes.

**Why this priority**: Nice-to-have enhancement that improves flexibility but isn't blocking current workflows. Users can currently launch clients directly with parameters if needed.

**Independent Test**: Can be tested by running `zen -- --verbose --debug` and verifying that the `--verbose --debug` flags are passed to the client at startup. Can also test priority by running `zen --yes -- --permission-mode acceptEdits` and verifying that `acceptEdits` is used (not `bypassPermissions`).

**Acceptance Scenarios**:

1. **Given** a user wants to pass client parameters, **When** they run `zen -- --client-flag value`, **Then** the system launches the client with `--client-flag value`
2. **Given** a user runs `zen use profile-name -- --client-flag value`, **When** the command executes, **Then** the system switches to the profile and launches the client with the specified flags
3. **Given** a user passes invalid client parameters, **When** the client starts, **Then** the client displays its own error message for invalid flags
4. **Given** a user runs `zen --yes -- --permission-mode acceptEdits`, **When** the client starts, **Then** it receives `--permission-mode acceptEdits` (not `bypassPermissions`), because `--` parameters have highest priority
5. **Given** a user has Web UI auto-permission enabled and runs `zen -- --permission-mode default`, **When** the client starts, **Then** it receives `--permission-mode default`, overriding the Web UI config

---

### Edge Cases

**Handled by Implementation**:
- Client binary not found: Display error "{client} not found in PATH", exit with code 1 (FR-012, FR-013)
- Permission flag detection: Use exact string matching for `--permission-mode` (Claude), `-a`/`--ask-for-approval` (Codex)
- Invalid permission mode in Web UI config: Log warning, fall back to default behavior, do not block client launch
- Special characters in `--` parameters: Preserve exactly as provided, client handles parsing
- Multiple permission flags in `--`: Pass all to client, client determines behavior
- Client doesn't support permission flags: Client will display its own error for unknown flags

**Requires Clarification During Implementation**:
- OpenCode permission behavior: Assume auto-approves by default, verify during T022
- Switching between clients with different flag formats: Each client receives its own format per configuration

## Requirements

### Functional Requirements

- **FR-001**: System MUST support a `--yes` / `-y` flag that passes `--permission-mode bypassPermissions` to Claude Code
- **FR-002**: System MUST map `--yes` flag to `-a never` for Codex
- **FR-003**: System MUST provide a Web UI configuration option to enable auto-permission mode per client (Claude Code, Codex, OpenCode)
- **FR-004**: Web UI MUST display client-specific permission options (Claude Code: 5 modes, Codex: 3 modes) based on selected client
- **FR-005**: System MUST store permission configuration separately for each client type
- **FR-006**: System MUST apply the configured auto-permission mode to all `zen` and `zen use` commands when enabled for the active client
- **FR-007**: System MUST implement priority order: `--` parameters > `--yes` flag > Web UI config > default behavior
- **FR-008**: System MUST support a `--` separator to pass arbitrary parameters to the underlying client
- **FR-009**: System MUST preserve the order and format of client parameters passed after `--` separator
- **FR-010**: System MUST NOT add `--yes` permission flags when `--` contains permission-related parameters
- **FR-011**: System MUST handle both `zen use profile-name -- [client-params]` and `zen -- [client-params]` patterns
- **FR-012**: System MUST display clear error messages when client launch fails
- **FR-013**: System MUST exit with appropriate status codes (0 for success, non-zero for errors) to support automation
- **FR-014**: System MUST document the new flags, priority order, and configuration options in help text and README
- **FR-015**: System MUST handle OpenCode client appropriately (determine if it has permission flags or auto-approves by default)

### Key Entities

- **Profile**: Represents a Claude API provider configuration with name, API key, base URL, and client preferences
- **Configuration**: Global settings including per-client auto-permission mode settings
  - `claude_auto_permission`: { enabled: boolean, mode: string } - Claude Code permission config
  - `codex_auto_permission`: { enabled: boolean, mode: string } - Codex permission config
  - `opencode_auto_permission`: { enabled: boolean, mode: string } - OpenCode permission config
- **Permission Mode**: Client-specific permission level
  - Claude Code: `default`, `acceptEdits`, `bypassPermissions`, `dontAsk`, `plan`
  - Codex: `untrusted`, `on-request`, `never` (via `-a` flag)
  - OpenCode: TBD (may auto-approve by default)
- **Client Parameters**: Command-line arguments passed through to the client after the `--` separator

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can run commands in automation scripts without manual intervention (0 interactive prompts with `--yes` flag)
- **SC-002**: Command execution with `--yes` flag completes in under 1 second for typical configurations
- **SC-003**: 100% of client parameters passed after `--` separator are correctly forwarded to the client
- **SC-004**: Users can configure auto-permission once and benefit from it across all subsequent command executions
- **SC-005**: Error scenarios (invalid client, missing config) provide clear feedback within 2 seconds

## Assumptions

- The current `zen` command already passes `--permission-mode acceptEdits` when `--yes` is used, but this needs to be changed to `bypassPermissions`
- Users are familiar with the `--` separator pattern from other CLI tools (e.g., `npm run script -- --flag`)
- The configuration file format supports adding new fields for auto-permission settings without breaking existing installations
- Client parameter validation is the responsibility of the client, not the zen wrapper
- Auto-permission setting applies globally to all profiles (not per-profile configuration)
- Claude Code supports `--permission-mode bypassPermissions` flag
- Codex has equivalent permission control flags that can be mapped to similar behavior
- OpenCode either has permission flags or auto-approves by default
