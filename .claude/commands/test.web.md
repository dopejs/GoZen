---
description: Run frontend tests with vitest and check coverage thresholds.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Goal

Run the frontend test suite using vitest with coverage reporting and verify thresholds are met.

## Steps

1. **Install dependencies** (if needed):
   ```bash
   cd web && pnpm install
   ```

2. **Run tests with coverage**:
   ```bash
   cd web && pnpm test -- --run --coverage
   ```

3. **Check coverage thresholds** (from vitest.config.ts):
   - Statements: **70%** minimum
   - Branches: **70%** minimum
   - Functions: **60%** minimum
   - Lines: **70%** minimum

4. **Report results**:
   - List each test file: PASS/FAIL with test names
   - Show page-level coverage (monitoring, providers, profiles, settings)
   - Show hook-level coverage (useProviders, useProfiles, etc.)
   - Flag any drops below thresholds
   - Show overall summary: X passed, Y failed, Z skipped

## Notes

- Tests use vitest + @testing-library/react + MSW for API mocking
- Test setup is in `web/src/test/setup.ts`
- If MSW handlers need updating, they're in `web/src/test/handlers.ts`
- Use `pnpm` (not npm) as the package manager
