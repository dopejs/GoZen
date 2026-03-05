# Contracts: Code Scenario Routing

No new external interfaces are introduced by this feature.

The `code` scenario uses the existing `routing` map structure in `zen.json` and the existing
Web API endpoints (`/api/v1/profiles`). No API contract changes are needed.

## Existing Contracts (unchanged)

- **Config schema**: `routing` map in `ProfileConfig` — already supports arbitrary `Scenario` string keys
- **Web API**: `GET/PUT /api/v1/profiles/:name` — profile structure unchanged
- **Proxy behavior**: Scenario detection + provider chain selection — extended with new scenario, same contract
