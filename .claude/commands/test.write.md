---
description: Analyze git diff and generate skeleton test cases following project TDD conventions.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Goal

Analyze code changes in the current branch and generate skeleton test cases that follow the project's existing test patterns and TDD conventions.

## Steps

1. **Identify changes** against the base branch:
   ```bash
   git diff main...HEAD --name-only | grep '\.go$\|\.tsx\?$'
   ```
   If user specifies files, use those instead.

2. **Analyze existing test patterns** in the codebase:
   - Go unit tests: table-driven tests in `*_test.go` files alongside source
   - Go integration tests: `test/integration/` with `//go:build integration` tag, `BaseTestConfig` embedding
   - Go e2e tests: `tests/` with `testEnv` struct, `MockProvider` helpers
   - Frontend tests: `*.test.tsx` with vitest + @testing-library/react + MSW

3. **For each changed file**, generate skeleton tests:

   **Go source files** (`internal/`, `cmd/`):
   - Create or update `*_test.go` in the same package
   - Use table-driven tests for functions with multiple input scenarios
   - Follow existing naming: `TestFunctionName_Scenario`
   - Include `t.Helper()` in test helpers

   **Go integration-worthy changes** (`internal/proxy/`, `internal/web/`, `internal/config/`):
   - Add integration test in `test/integration/` using `MockProvider`
   - Use the correct test config type: `ProxyTestConfig`, `WebTestConfig`, or `DaemonTestConfig`

   **Frontend files** (`web/src/pages/`, `web/src/hooks/`):
   - Create or update `*.test.tsx` alongside the component
   - Use `render()`, `screen`, `userEvent` from @testing-library/react
   - Mock API calls with MSW handlers

4. **Output skeleton tests** with:
   - Test function signatures and descriptions
   - `// TODO: implement` markers for test bodies
   - Required imports and setup
   - Suggested assertions based on the code being tested

5. **Report coverage gaps**:
   - List functions/components without tests
   - Suggest which tests would provide the most value

## Notes

- Generate skeletons only — do not write full test implementations unless the user asks
- Prefer extending existing test files over creating new ones
- Follow the project's `CLAUDE.md` rule: "Minimal test files — only add tests for new public APIs or complex logic"
- Table-driven tests are preferred for Go code
