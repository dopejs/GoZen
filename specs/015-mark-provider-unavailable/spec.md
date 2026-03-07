# Feature Specification: Manual Provider Unavailability Marking

**Feature Branch**: `015-mark-provider-unavailable`
**Created**: 2026-03-07
**Status**: Draft
**Input**: User description: "手动标注一个 provider 为不可用，支持 Web UI 和 CLI 两种方式，分今日/本月/永久三种有效期，标注后 proxy 自动跳过该 provider"

## Clarifications

### Session 2026-03-07

- Q: When all providers in a profile are marked unavailable and the "last resort" rule kicks in, how should the system handle provider selection? → A: Return an error directly. Do not attempt any provider. The marking is often a cost-related decision, so the error prompts the user to consciously decide which provider to re-enable.
- Q: When a scenario route falls back to the profile's default provider list, should unavailability markings on the default list providers still be enforced? → A: Yes. Markings are enforced on the fallback list too. If all default providers are also unavailable, return an error. This is consistent with cost-control intent.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Mark Provider Unavailable via Web UI (Priority: P1)

A user discovers that one of their API providers is experiencing issues (e.g., billing suspended, rate limited, maintenance window). They open the GoZen Web UI, navigate to the provider list or health monitoring page, and mark that provider as unavailable. They select an expiration period — today only, this calendar month, or permanently. The system immediately stops routing requests to that provider.

**Why this priority**: This is the core user interaction. The Web UI is the primary management interface for most users and provides the most intuitive way to manage provider availability with visual feedback.

**Independent Test**: Can be fully tested by opening the Web UI, marking a provider unavailable, then verifying proxy requests skip that provider. Delivers immediate value by preventing wasted requests to known-bad providers.

**Acceptance Scenarios**:

1. **Given** a profile with 3 healthy providers (A, B, C), **When** the user marks provider B as "unavailable today" via Web UI, **Then** all subsequent proxy requests skip provider B and route to A or C only.
2. **Given** a provider marked as "unavailable today", **When** the next calendar day begins (local midnight), **Then** the provider automatically becomes available again without user action.
3. **Given** a provider marked as "unavailable this month", **When** the next calendar month begins, **Then** the provider automatically becomes available again.
4. **Given** a provider marked as "permanently unavailable", **When** time passes, **Then** the provider remains unavailable until the user manually clears the marking.
5. **Given** a provider currently marked unavailable, **When** the user clears the unavailability marking via Web UI, **Then** the provider immediately becomes available for routing again.

---

### User Story 2 - Mark Provider Unavailable via CLI (Priority: P1)

A user wants to quickly mark a provider as unavailable from the terminal without opening the Web UI. They run a CLI command specifying the provider name and the unavailability duration. They can also list currently marked providers and clear markings from the command line.

**Why this priority**: CLI access is equally important as Web UI for power users and automation scripts. Many GoZen users work primarily in the terminal.

**Independent Test**: Can be fully tested by running the CLI command to mark a provider, then verifying via `zen config` or proxy behavior that the provider is skipped. Delivers value for terminal-centric workflows and scriptable automation.

**Acceptance Scenarios**:

1. **Given** a configured provider named "openai-backup", **When** the user runs the CLI command to mark it unavailable for today, **Then** the provider is immediately marked and proxy requests skip it.
2. **Given** multiple providers with various unavailability markings, **When** the user runs a list/status command, **Then** they see all marked providers with their expiration type and remaining duration.
3. **Given** a provider marked as unavailable, **When** the user runs the CLI command to clear the marking, **Then** the provider is immediately available for routing.
4. **Given** a nonexistent provider name, **When** the user runs the mark command, **Then** an error message is shown indicating the provider does not exist.

---

### User Story 3 - Error on All-Unavailable and Scenario Fallback (Priority: P1)

When all providers in a profile are marked unavailable, the system returns an error to the user instead of attempting any provider. This is intentional — markings often reflect cost decisions, and the error prompts the user to consciously decide which provider to re-enable. For scenario routes, when all route-specific providers are marked unavailable, the system falls back to the profile's default provider list before potentially returning an error.

**Why this priority**: This defines the critical boundary behavior. Returning a clear error instead of silently using a cost-prohibitive provider respects the user's intent and avoids unexpected charges.

**Independent Test**: Can be tested by marking all providers in a profile as unavailable and verifying the proxy returns an error response with a clear message explaining that all providers are unavailable.

**Acceptance Scenarios**:

1. **Given** a profile with only one provider, **When** that provider is marked unavailable, **Then** the proxy returns an error indicating all providers are unavailable and the user should re-enable one.
2. **Given** a profile with providers A, B, C, **When** all three are marked unavailable, **Then** the proxy returns an error with a message listing the unavailable providers and suggesting the user clear a marking.
3. **Given** a scenario route with providers X and Y where both are marked unavailable, **When** a request matches that scenario, **Then** the system falls back to the profile's default provider list.
4. **Given** a scenario route with a single provider Z marked unavailable, **When** a request matches that scenario, **Then** the system falls back to the profile's default provider list.
5. **Given** a scenario route where all providers are unavailable AND the profile's default list also has all providers unavailable, **When** a request arrives, **Then** the proxy returns an error indicating all providers are unavailable.

---

### User Story 4 - Direct Provider Selection Bypasses Unavailability (Priority: P2)

When a user explicitly selects a specific provider using `zen use <provider>`, the unavailability marking is ignored. The user's explicit intent to use that provider overrides the marking.

**Why this priority**: Users who explicitly choose a provider have a reason for doing so. Blocking their explicit choice would be frustrating and counterproductive.

**Independent Test**: Can be tested by marking a provider unavailable, then running `zen use <provider>` and verifying the provider is used regardless.

**Acceptance Scenarios**:

1. **Given** provider "my-anthropic" is marked as permanently unavailable, **When** the user runs `zen use my-anthropic`, **Then** the provider is used directly without any warning or blocking.
2. **Given** provider "backup-api" is marked as unavailable today, **When** the user runs `zen use backup-api`, **Then** the client launches using that provider normally.

---

### User Story 5 - Visibility of Unavailability Status (Priority: P2)

Users can see which providers are currently marked unavailable, with clear indicators in both the Web UI and CLI output. The health monitoring page reflects manual unavailability distinctly from automatic health detection.

**Why this priority**: Without clear visibility, users may forget which providers they've marked and be confused by routing behavior.

**Independent Test**: Can be tested by marking providers and checking the Web UI provider list and health page for visual indicators, and CLI status output for text indicators.

**Acceptance Scenarios**:

1. **Given** a provider marked as unavailable, **When** the user views the provider list in Web UI, **Then** the provider shows a distinct visual indicator (e.g., badge or icon) showing its unavailability status and expiration type.
2. **Given** a provider marked as unavailable, **When** the user views the health monitoring page, **Then** the manual unavailability status is shown separately from automatic health detection status.
3. **Given** providers with different expiration types, **When** the user views the provider list, **Then** each provider shows its specific expiration type (today, this month, permanent) and when it will auto-expire (if applicable).

---

### Edge Cases

- What happens when a provider is marked unavailable and a config reload occurs? The marking must persist across reloads.
- What happens when a provider is renamed or deleted while it has an unavailability marking? The marking should be cleaned up.
- What happens when the system clock changes (e.g., timezone change, NTP correction) while "today" or "this month" markings are active? Expiration should be based on the local system clock at evaluation time.
- What happens when the daemon restarts? Unavailability markings must persist across daemon restarts.
- What happens when a provider is marked unavailable via CLI while the daemon is not running? The marking should be saved to config and take effect when the daemon next starts.
- What happens when all providers in a profile's default list AND all scenario routes are marked unavailable? The proxy returns an error response; no provider is attempted.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow users to mark any configured provider as unavailable through the Web UI.
- **FR-002**: System MUST allow users to mark any configured provider as unavailable through a CLI command.
- **FR-003**: System MUST support three unavailability duration types: today (until end of current calendar day in local time), this month (until end of current calendar month in local time), and permanent (until manually cleared).
- **FR-004**: System MUST allow users to clear an unavailability marking at any time through both Web UI and CLI, restoring the provider to available status immediately.
- **FR-005**: System MUST automatically expire "today" markings at the start of the next calendar day (local time) and "this month" markings at the start of the next calendar month (local time).
- **FR-006**: System MUST skip providers marked as unavailable (and within their validity period) during proxy request routing, proceeding to the next available provider in the chain.
- **FR-007**: System MUST return an error response when all providers in the active profile's provider list are marked as unavailable. The error message MUST clearly indicate that all providers are unavailable and suggest the user re-enable a provider. The system MUST NOT silently attempt any marked-unavailable provider.
- **FR-008**: System MUST fall back to the profile's default provider list when all providers in a scenario route are marked unavailable. Unavailability markings MUST still be enforced on the fallback list — if all default providers are also marked unavailable, the system MUST return an error per FR-007.
- **FR-009**: System MUST NOT enforce unavailability markings when a user explicitly selects a provider via `zen use <provider>`.
- **FR-010**: System MUST persist unavailability markings across daemon restarts and config reloads.
- **FR-011**: System MUST display unavailability status (including expiration type) in the Web UI provider list and health monitoring views.
- **FR-012**: System MUST display unavailability status in CLI provider listing and status commands.
- **FR-013**: System MUST clean up unavailability markings for providers that are deleted from the configuration.
- **FR-014**: System MUST distinguish between manual unavailability markings and automatic health-based unavailability in all status displays.

### Key Entities

- **Unavailability Marking**: Represents a user's manual decision to mark a provider as unavailable. Key attributes: target provider name, expiration type (today/month/permanent), creation timestamp, expiration timestamp (calculated from type). Associated with exactly one provider. Multiple markings on the same provider are not allowed — a new marking replaces any existing one.
- **Provider (extended)**: The existing provider entity gains an optional unavailability marking. The provider's effective availability is determined by combining automatic health status with any manual unavailability marking.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can mark a provider as unavailable in under 10 seconds via either Web UI or CLI.
- **SC-002**: Proxy requests never attempt a provider that is manually marked unavailable. When all providers are marked unavailable, the proxy returns an error instead of attempting any, resulting in zero wasted request attempts to known-unavailable providers.
- **SC-003**: Expired "today" and "this month" markings are automatically cleared within 1 minute of the expiration boundary, requiring no user intervention.
- **SC-004**: 100% of unavailability markings persist correctly across daemon restarts with no data loss.
- **SC-005**: Users can identify all manually unavailable providers and their expiration status at a glance in both Web UI and CLI output.
- **SC-006**: Direct provider selection via `zen use` works without any interference from unavailability markings, maintaining 100% backward compatibility.

## Assumptions

- **Time zone handling**: "Today" and "this month" expiration boundaries are based on the server's local system clock and timezone, not UTC. This is consistent with user expectations for a locally-running tool.
- **Marking granularity**: A provider can have at most one unavailability marking at a time. Setting a new marking replaces any existing one on the same provider.
- **No confirmation dialog**: Marking a provider unavailable does not require a confirmation dialog, as the action is easily reversible. Clearing an unavailability marking also requires no confirmation.
- **Daemon communication**: When the CLI marks a provider unavailable while the daemon is running, the daemon picks up the change via config reload. When the daemon is not running, the marking is persisted to config and takes effect on next start.
- **Monitoring integration**: The existing request monitoring system logs which providers were skipped due to manual unavailability, providing audit trail visibility.
