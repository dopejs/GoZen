# Feature Specification: Provider Proxy Settings

**Feature Branch**: `001-provider-proxy`
**Created**: 2026-02-27
**Status**: Draft
**Input**: User description: "为provider加上一个proxy的设置，可以选择proxy的类型为http, https, socks5，zen通过proxy连接provider的baseurl。要考虑 zen use 这种用法，是不经过我们的代理服务器的"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure a proxy for a provider (Priority: P1)

A user is behind a corporate firewall or in a region where direct access
to the provider's API endpoint is blocked. They need to route traffic
through a proxy server to reach the provider. The user adds or edits a
provider configuration and specifies a proxy URL with a protocol type
(HTTP, HTTPS, or SOCKS5). Once saved, all API requests that GoZen's
daemon proxy makes to that provider's base URL go through the configured
proxy.

**Why this priority**: This is the core value of the feature — without
it, users in restricted networks cannot use GoZen at all.

**Independent Test**: Can be fully tested by configuring a provider with
a proxy URL, running `zen` with that provider, and verifying that
outbound traffic passes through the specified proxy.

**Acceptance Scenarios**:

1. **Given** a provider with `proxy_url` set to `http://proxy.corp:8080`,
   **When** GoZen's daemon sends a request to that provider's base URL,
   **Then** the request is routed through `http://proxy.corp:8080`.

2. **Given** a provider with `proxy_url` set to `socks5://socks.local:1080`,
   **When** GoZen's daemon sends a request to that provider's base URL,
   **Then** the request is routed through the SOCKS5 proxy.

3. **Given** a provider with no `proxy_url` configured,
   **When** GoZen's daemon sends a request to that provider's base URL,
   **Then** the request is sent directly (current behavior, unchanged).

4. **Given** a provider with an invalid proxy URL (e.g., malformed
   address or unsupported scheme),
   **When** the user saves the provider configuration,
   **Then** the system rejects the configuration with a clear error
   message indicating the problem.

---

### User Story 2 - Direct connection with `zen use` (Priority: P2)

A user runs `zen use <provider>` which bypasses GoZen's daemon proxy
and directly exec's the CLI with environment variables set. When using
this mode, the CLI (Claude Code, Codex, OpenCode) connects directly to
the provider's base URL. If the provider has a proxy configured, GoZen
MUST set the appropriate environment variables (`HTTP_PROXY`,
`HTTPS_PROXY`, `ALL_PROXY`) so that the CLI process inherits the proxy
setting and routes its own traffic through the configured proxy.

**Why this priority**: `zen use` is a supported usage path that skips
the daemon proxy entirely. Without proxy env var propagation, users who
depend on a proxy would have broken connectivity in this mode.

**Independent Test**: Can be tested by configuring a provider with a
proxy, running `zen use <provider>`, and verifying the spawned CLI
process has the correct `HTTP_PROXY`/`HTTPS_PROXY`/`ALL_PROXY`
environment variables set.

**Acceptance Scenarios**:

1. **Given** a provider with `proxy_url` set to `http://proxy.corp:8080`,
   **When** the user runs `zen use <provider>`,
   **Then** the spawned CLI process has `HTTP_PROXY` and `HTTPS_PROXY`
   set to `http://proxy.corp:8080`.

2. **Given** a provider with `proxy_url` set to `socks5://socks.local:1080`,
   **When** the user runs `zen use <provider>`,
   **Then** the spawned CLI process has `ALL_PROXY` set to
   `socks5://socks.local:1080`.

3. **Given** a provider with no `proxy_url`,
   **When** the user runs `zen use <provider>`,
   **Then** no proxy environment variables are set (current behavior).

---

### User Story 3 - Manage proxy settings via TUI and Web UI (Priority: P3)

A user manages provider configurations through GoZen's TUI editor or
Web UI. The proxy field appears as an optional setting when adding or
editing a provider. The user can input a proxy URL and the system
validates the scheme (must be http, https, or socks5) and format before
saving.

**Why this priority**: UI support improves the user experience but the
feature works without it — users can edit `zen.json` directly.

**Independent Test**: Can be tested by opening the TUI/Web UI provider
editor, entering a proxy URL, saving, and verifying it appears in
`zen.json`.

**Acceptance Scenarios**:

1. **Given** the user is editing a provider in the TUI,
   **When** they enter a proxy URL like `socks5://proxy:1080`,
   **Then** the configuration is saved with the proxy field and the
   value is validated for correct scheme.

2. **Given** the user is editing a provider in the Web UI,
   **When** they enter a proxy URL in the proxy field,
   **Then** the configuration is saved and subsequent API requests to
   that provider use the proxy.

3. **Given** the user enters a proxy URL with an unsupported scheme
   (e.g., `ftp://proxy:21`),
   **When** they attempt to save,
   **Then** validation rejects the input with a message listing the
   supported schemes.

---

### Edge Cases

- What happens when the proxy server is unreachable? The system MUST
  treat it like a provider failure — failover to the next provider in
  the profile (if available) and log the error.
- What happens when a proxy URL uses an IP address instead of a
  hostname? It MUST be accepted and work correctly.
- What happens when multiple providers in a profile have different
  proxies? Each provider's requests MUST use its own configured proxy
  independently.
- What happens when the proxy requires authentication
  (e.g., `http://user:pass@proxy:8080`)? The system MUST support
  embedded credentials in the proxy URL.
- What happens during config migration from a version without the
  proxy field? Existing configurations MUST load without error, with
  the proxy field defaulting to empty (no proxy).

## Clarifications

### Session 2026-02-28

- Q: Should proxy_url be included in config sync? → A: Exclude proxy_url from sync; users configure proxy separately on each device.
- Q: Should proxy usage be logged in structured logs? → A: Log proxy URL (credentials masked) in existing structured request logs.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Each provider configuration MUST support an optional
  `proxy_url` field that specifies the proxy server to use for
  connecting to that provider's base URL.
- **FR-002**: The system MUST support three proxy schemes: `http`,
  `https`, and `socks5`.
- **FR-003**: When `proxy_url` is configured for a provider, all
  outbound requests from GoZen's daemon proxy to that provider's
  base URL MUST be routed through the specified proxy.
- **FR-004**: When `proxy_url` is configured and the user runs
  `zen use <provider>`, the system MUST set appropriate proxy
  environment variables (`HTTP_PROXY`, `HTTPS_PROXY`, `ALL_PROXY`)
  on the spawned CLI process.
- **FR-005**: The system MUST validate that `proxy_url` uses a
  supported scheme (`http`, `https`, or `socks5`) when saving
  configuration.
- **FR-006**: The system MUST validate that `proxy_url` is a
  well-formed URL when saving configuration.
- **FR-007**: Providers without `proxy_url` MUST continue to connect
  directly to their base URL (no behavior change).
- **FR-008**: The proxy setting MUST be per-provider — different
  providers in the same profile MAY use different proxies.
- **FR-009**: The TUI provider editor MUST include a proxy URL field.
- **FR-010**: The Web UI provider editor MUST include a proxy URL field.
- **FR-011**: Existing configurations without the proxy field MUST
  load successfully with the proxy defaulting to empty.
- **FR-012**: Proxy URLs with embedded authentication credentials
  MUST be supported (e.g., `http://user:pass@host:port`).
- **FR-013**: Provider health checks MUST also route through the
  configured proxy for that provider.
- **FR-014**: The `proxy_url` field MUST be excluded from config sync.
  Users MUST configure proxy settings independently on each device.
- **FR-015**: Structured request logs MUST include the proxy URL used
  for each request (with credentials masked, e.g.,
  `http://***@proxy:8080`). Proxy connection errors MUST also be
  logged with masked URLs.

### Key Entities

- **Provider Configuration**: Extended with an optional proxy URL
  (scheme, host, port, optional credentials). Related to existing
  provider fields (base_url, auth_token, models, env_vars).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can configure a proxy for a provider and
  successfully connect to the provider's API through that proxy on the
  first attempt, with no manual workaround required.
- **SC-002**: 100% of existing configurations (without proxy field)
  load and function identically after the update — zero regressions.
- **SC-003**: `zen use <provider>` with a proxy-configured provider
  results in the spawned CLI process having correct proxy environment
  variables, verified by inspecting the process environment.
- **SC-004**: All three proxy schemes (HTTP, HTTPS, SOCKS5) work
  correctly when tested against a proxy server of each type.
- **SC-005**: Invalid proxy URLs are rejected at configuration time
  with a clear, actionable error message — users do not encounter
  cryptic runtime failures.

## Assumptions

- Proxy authentication, when needed, is provided via embedded
  credentials in the URL (standard `user:pass@host` format) rather
  than a separate authentication mechanism.
- SOCKS5 support includes both SOCKS5 and SOCKS5h (DNS resolution
  at the proxy). The default `socks5` scheme resolves DNS at the
  proxy (SOCKS5h behavior), which is the most common expectation for
  users behind restrictive firewalls.
- The proxy setting applies only to the connection between GoZen and
  the upstream provider — it does not affect connections between the
  local CLI and GoZen's local daemon proxy (which is always on
  127.0.0.1).
