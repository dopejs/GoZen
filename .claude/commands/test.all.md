---
description: Run all test tiers (unit, integration, e2e, web) and produce a consolidated summary.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Goal

Run the complete test pyramid — unit tests, integration tests, e2e tests, and frontend tests — in sequence, and produce a consolidated pass/fail summary.

## Steps

1. **Go unit tests**:
   ```bash
   go test -race -count=1 -coverprofile=coverage.out ./...
   ```

2. **Build binary for integration/e2e tests**:
   ```bash
   cd web && pnpm install && pnpm build && cd ..
   go build -o bin/zen .
   ```

3. **Go integration tests**:
   ```bash
   go test -tags=integration -v -timeout 120s ./test/integration/...
   ```

4. **Go e2e tests**:
   ```bash
   go test -tags=integration -v -timeout 180s ./tests/...
   ```

5. **Frontend tests**:
   ```bash
   cd web && pnpm test -- --run --coverage
   ```

6. **Consolidated report**:
   ```
   === Test Summary ===
   Unit tests:        X passed, Y failed
   Integration tests: X passed, Y failed
   E2E tests:         X passed, Y failed
   Frontend tests:    X passed, Y failed

   Coverage:
     Go (core):       XX% (threshold: 80%)
     Go (supporting): XX% (threshold: 50%)
     Frontend:        XX% (threshold: 70%)

   Result: PASS / FAIL
   ```

7. **On full pass**, suggest running `/commit` to commit changes.

## Notes

- Tests run in sequence: unit → integration → e2e → frontend
- If unit tests fail, still continue with remaining tiers to get full picture
- The e2e stress test sends 500+ requests and may add significant time
- Use `pnpm` (not npm) for frontend tests
