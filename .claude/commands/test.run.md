---
description: Run Go unit tests with race detection and coverage for modified packages.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Goal

Run Go unit tests with race detection and coverage reporting. Detect which packages have been modified and focus on those, or run all tests if no specific packages are identified.

## Steps

1. **Detect modified packages** (if no user input specifies packages):
   ```bash
   git diff --name-only HEAD | grep '\.go$' | xargs -I{} dirname {} | sort -u
   ```
   If no modified Go files, run all packages.

2. **Run tests with race detection and coverage**:
   ```bash
   go test -race -count=1 -coverprofile=coverage.out ./...
   ```
   Or for specific packages:
   ```bash
   go test -race -count=1 -coverprofile=coverage.out ./path/to/package/...
   ```

3. **Check coverage against thresholds**:
   - Core packages (`internal/proxy`, `internal/config`, `internal/web`): **80%** minimum
   - Supporting packages (`cmd/`, `tui/`, `internal/daemon`, `internal/update`): **50%** minimum

4. **Report results**:
   - List each test: PASS/FAIL with test name
   - Show coverage per package
   - Flag any packages below threshold
   - Show overall summary: X passed, Y failed, Z skipped

## Notes

- Do NOT run integration tests (those use `-tags=integration`)
- If tests fail, show the failure output and suggest fixes
- Use `-count=1` to disable test caching
