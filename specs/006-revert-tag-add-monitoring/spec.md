# Feature Specification: Revert Provider Tag & Add Request Monitoring UI

**Feature Branch**: `006-revert-tag-add-monitoring`
**Created**: 2026-03-03
**Status**: Draft
**Input**: User description: "撤销 005-provider-model-tag 功能（该功能通过在响应内容中注入 provider tag 导致了持久化污染和 Bedrock API 错误），并实现 Web UI 实时请求监控页面，让用户可以在浏览器中查看每个请求使用的 provider、model、耗时、token 等信息，而不修改 API 响应内容。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Remove Provider Tag Injection (Priority: P1)

A developer using GoZen proxy with Bedrock encounters API errors (`Invalid 'signature' in 'thinking' block`) caused by the provider tag feature injecting text into thinking blocks. Additionally, even after disabling the feature or reverting to an older version, the injected tags persist in conversation history, causing ongoing issues. The provider tag injection code must be completely removed to stop polluting API responses and prevent future errors.

**Why this priority**: This is critical because the current implementation causes API failures and irreversible data pollution. Removing it is a prerequisite for system stability.

**Independent Test**: After code removal, send requests with thinking blocks enabled through the proxy and verify that responses are unmodified (no provider tags injected), and Bedrock API accepts the responses without validation errors.

**Acceptance Scenarios**:

1. **Given** the provider tag code has been removed, **When** a request with thinking block is proxied, **Then** the response contains no injected provider tag and Bedrock API returns 200 OK
2. **Given** the provider tag code has been removed, **When** any API request is proxied, **Then** the response body is identical to the upstream provider's response (byte-for-byte match)
3. **Given** the provider tag config field exists in user configs, **When** the system loads the config, **Then** the field is ignored (no errors, backward compatible)

---

### User Story 2 - Real-Time Request Monitoring Page (Priority: P1)

A developer using GoZen wants to know which provider and model handled each request, along with performance metrics like response time and token usage. They open the Web UI and navigate to the "Requests" page that displays a live feed of recent requests, showing provider name, model, timestamp, duration, token counts, and cost. This provides visibility without modifying API responses.

**Why this priority**: This delivers the original value proposition (visibility into provider usage) without the data pollution issues. It's the core replacement for the removed feature.

**Independent Test**: Start the daemon, make several API requests through the proxy, open the Web UI Requests page, and verify that all requests appear with correct provider, model, timing, and token information.

**Acceptance Scenarios**:

1. **Given** the Web UI Requests page is open, **When** a new request is proxied, **Then** the request appears in the monitoring feed within 1 second with provider, model, timestamp, and duration
2. **Given** multiple requests have been made, **When** the user views the Requests page, **Then** requests are displayed in reverse chronological order (newest first) with all metadata visible
3. **Given** the Requests page is displaying requests, **When** the user refreshes the page, **Then** the request history persists and displays the same data

---

### User Story 3 - Request Detail View (Priority: P2)

A developer investigating a specific request wants to see detailed information beyond the summary. They click on a request in the monitoring feed and see expanded details including request/response headers, token breakdown (input/output/cache), error messages (if any), and failover history (which providers were tried before success).

**Why this priority**: Provides deeper debugging capabilities for troubleshooting issues. Lower priority because the summary view (P1) already delivers core value.

**Independent Test**: Click on any request in the monitoring feed and verify that a detail panel or modal opens showing comprehensive request metadata.

**Acceptance Scenarios**:

1. **Given** a request is displayed in the monitoring feed, **When** the user clicks on it, **Then** a detail view opens showing full request metadata including headers, token breakdown, and timing details
2. **Given** a request failed over to a backup provider, **When** the user views the request details, **Then** the failover history shows which providers were attempted and why each failed
3. **Given** a request resulted in an error, **When** the user views the request details, **Then** the error message and status code are clearly displayed

---

### User Story 4 - Request Filtering and Search (Priority: P3)

A developer wants to analyze requests by specific criteria (e.g., all requests to a specific provider, all errors, all requests using a specific model). They use filter controls on the Requests page to narrow down the request list by provider, model, status code, or time range.

**Why this priority**: Enhances usability for power users but not essential for basic monitoring. Can be added incrementally.

**Independent Test**: Apply various filters (provider, model, status) and verify that the request list updates to show only matching requests.

**Acceptance Scenarios**:

1. **Given** the Requests page displays multiple requests, **When** the user filters by a specific provider, **Then** only requests handled by that provider are shown
2. **Given** the Requests page displays multiple requests, **When** the user filters by error status, **Then** only failed requests (4xx/5xx) are shown
3. **Given** the Requests page displays multiple requests, **When** the user sets a time range filter, **Then** only requests within that time range are shown

---

### Edge Cases

- What happens when the daemon restarts? → Request history is lost (in-memory storage for MVP, can be persisted later)
- What happens when thousands of requests accumulate? → Implement a rolling buffer (e.g., keep last 1000 requests) to prevent memory issues
- What happens when a request fails on all providers? → Display the request with error status and show all attempted providers in detail view
- What happens when token usage is not available (streaming responses)? → Display "N/A" or "streaming" for token fields
- What happens when the Web UI is not open? → Requests are still logged in memory (up to buffer limit) and persisted to SQLite, visible when UI is opened
- What happens with concurrent requests? → Each request gets a unique ID and timestamp, displayed in order of completion

## Requirements *(mandatory)*

### Functional Requirements

#### Removal Requirements

- **FR-001**: System MUST remove all provider tag injection code from the proxy response handling logic
- **FR-002**: System MUST remove the `show_provider_tag` configuration field from the config schema
- **FR-003**: System MUST remove the provider tag toggle from the Web UI settings page
- **FR-004**: System MUST maintain backward compatibility when loading configs that contain the deprecated `show_provider_tag` field (parse and silently ignore without errors, field can be removed from struct using custom UnmarshalJSON)
- **FR-005**: System MUST NOT modify API response bodies in any way (responses must be byte-for-byte identical to upstream provider responses, except for format transformation when needed)

#### Monitoring Requirements

- **FR-006**: System MUST capture metadata for each proxied request as defined in the RequestRecord schema (see data-model.md)
- **FR-007**: System MUST capture token usage (input tokens, output tokens, cache tokens) when available in the response
- **FR-008**: System MUST store request metadata in memory with a configurable buffer size (default: 1000 requests)
- **FR-009**: System MUST provide a REST API endpoint to retrieve recent request history (e.g., `GET /api/v1/monitoring/requests`)
- **FR-010**: Web UI MUST include a new "Requests" page accessible from the main navigation
- **FR-011**: The Requests page MUST display requests in reverse chronological order with columns for: timestamp, provider, model, status, duration, tokens, cost
- **FR-012**: The Requests page MUST auto-refresh or support manual refresh to show new requests
- **FR-013**: System MUST calculate and display estimated cost per request based on model pricing
- **FR-014**: For failed requests, system MUST capture and display error messages and status codes
- **FR-015**: For requests that failed over to backup providers, system MUST capture the failover history

### Key Entities

- **Request Record**: Represents a single proxied API request with metadata including provider, model, timing, tokens, cost, status, and error information. Stored in memory with a rolling buffer to prevent unbounded growth. Implemented as `RequestRecord` struct in code.

- **Failover History**: For requests that tried multiple providers before succeeding (or failing), tracks which providers were attempted, in what order, and why each failed (error message, status code).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After removal, 100% of API responses are unmodified by the proxy (excluding necessary format transformations)
- **SC-002**: After removal, 0% of requests with thinking blocks result in Bedrock validation errors caused by injected content
- **SC-003**: Users can view request history for the last 1000 requests in the Web UI with API response time <500ms (localhost) and total time-to-interactive <2 seconds
- **SC-004**: New requests appear in the Requests page within 1 second of completion
- **SC-005**: The Requests page displays all key metrics (provider, model, duration, tokens, cost) for each request without requiring additional clicks
- **SC-006**: Users can identify which provider handled a specific request by viewing the Requests page, achieving the original goal without response modification

## Assumptions

- Request metadata storage uses in-memory ring buffer (1000 requests) as primary storage, with optional asynchronous SQLite persistence for durability across daemon restarts
- The existing structured logger and log database can be leveraged or extended for request tracking
- Token usage information is available in non-streaming responses; streaming responses will store 0 for token fields (displayed as "N/A" or "-" in Web UI)
- Cost calculation uses existing pricing data from the usage tracking system
- The monitoring page is a new top-level page in the Web UI (titled "Requests"), not a sub-tab of settings
- Real-time updates use polling (HTTP requests every few seconds) rather than WebSockets for simplicity
- The buffer size of 1000 requests is sufficient for typical usage patterns (can be made configurable later)
- SQLite persistence is non-blocking and does not impact request latency (writes happen asynchronously)
