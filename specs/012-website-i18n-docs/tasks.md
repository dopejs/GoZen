# Implementation Tasks: Website Documentation Internationalization

**Feature**: 012-website-i18n-docs
**Date**: 2026-03-06
**Based on**: [spec.md](./spec.md), [plan.md](./plan.md), [data-model.md](./data-model.md), [contracts/cli-commands.md](./contracts/cli-commands.md)

## Task Summary

- **Total Tasks**: 42
- **Phases**: 6 (Setup, Foundational, US1, US2, US3, Polish)
- **Parallel Opportunities**: 18 tasks marked [P]
- **Test-First Approach**: Tests written before implementation for all core logic

## Implementation Strategy

**MVP Scope**: User Story 1 (P1) - Identify Missing Translations
- Delivers immediate value: audit tool that shows what's missing
- Foundation for P2 and P3 features
- Can be released independently

**Incremental Delivery**:
1. Phase 1-3: MVP (audit command with table output)
2. Phase 4: Translation workflow support (list-missing, priority sorting)
3. Phase 5: Maintenance features (sync-check, staleness detection)
4. Phase 6: Polish (JSON/markdown output, CI/CD examples)

---

## Phase 1: Setup

**Goal**: Initialize project structure and dependencies

- [x] T001 Create project directory structure at website/tools/i18n-audit/
- [x] T002 Initialize Go module with `go mod init github.com/dopejs/gozen/website/tools/i18n-audit`
- [x] T003 Add dependencies: github.com/yuin/goldmark, gopkg.in/yaml.v3, github.com/spf13/cobra, github.com/charmbracelet/lipgloss
- [x] T004 Create main.go entry point with Cobra root command in website/tools/i18n-audit/main.go
- [x] T005 Create testdata/ directory with sample docs structure in website/tools/i18n-audit/testdata/

---

## Phase 2: Foundational

**Goal**: Core types and interfaces that all user stories depend on

- [x] T006 [P] Define DocumentationPage struct in website/tools/i18n-audit/types.go
- [x] T007 [P] Define Translation struct with TranslationStatus enum in website/tools/i18n-audit/types.go
- [x] T008 [P] Define Locale struct in website/tools/i18n-audit/types.go
- [x] T009 [P] Define AuditReport and LocaleReport structs in website/tools/i18n-audit/types.go
- [x] T010 [P] Define ExclusionRule struct in website/tools/i18n-audit/types.go
- [x] T011 Create test fixtures in website/tools/i18n-audit/testdata/sample-docs/ (English source files)
- [x] T012 Create test fixtures in website/tools/i18n-audit/testdata/sample-i18n/ (partial translations)

---

## Phase 3: User Story 1 - Identify Missing Translations (P1)

**Goal**: Implement audit command that scans docs and generates coverage report

**Independent Test**: Run `go run . audit` in website/tools/i18n-audit/ and verify it outputs a table showing translation status per locale with accurate missing file counts.

### Tests

- [x] T013 [P] [US1] Write scanner tests in website/tools/i18n-audit/scanner_test.go (table-driven: valid paths, empty dirs, invalid paths)
- [x] T014 [P] [US1] Write analyzer tests in website/tools/i18n-audit/analyzer_test.go (coverage calculation, missing detection, exclusion rules)
- [x] T015 [P] [US1] Write reporter tests in website/tools/i18n-audit/reporter_test.go (table format, coverage percentage, locale sorting)

### Implementation

- [x] T016 [US1] Implement ScanSourceDocs() in website/tools/i18n-audit/scanner.go (filepath.Walk, .md/.mdx filter, frontmatter parsing)
- [x] T017 [US1] Implement ScanTranslations() in website/tools/i18n-audit/scanner.go (detect locales, map translated files)
- [x] T018 [US1] Implement AnalyzeCoverage() in website/tools/i18n-audit/analyzer.go (compare source vs translations, calculate percentages)
- [x] T019 [US1] Implement DetectMissingFiles() in website/tools/i18n-audit/analyzer.go (identify missing translations per locale)
- [x] T020 [US1] Implement GenerateTableReport() in website/tools/i18n-audit/reporter.go (lipgloss table with locale rows)
- [x] T021 [US1] Implement audit command in website/tools/i18n-audit/cmd_audit.go (wire scanner → analyzer → reporter)
- [x] T022 [US1] Add --docs-path and --i18n-path flags to audit command in website/tools/i18n-audit/cmd_audit.go
- [x] T023 [US1] Add --locale flag to filter specific locale in website/tools/i18n-audit/cmd_audit.go

### Integration

- [x] T024 [US1] Test audit command against real website/docs/ directory
- [x] T025 [US1] Verify coverage calculation matches manual count (9/16 for zh-Hans)
- [x] T026 [US1] Verify missing files list matches expected 7 files (agent-infrastructure.md, bot.md, etc.)

---

## Phase 4: User Story 2 - Translation Workflow Support (P2)

**Goal**: Help maintainers prioritize and create translations

**Independent Test**: Run `go run . list-missing --locale zh-Hans --priority` and verify it outputs missing files sorted by size with suggested file paths.

### Tests

- [x] T027 [P] [US2] Write priority sorting tests in website/tools/i18n-audit/analyzer_test.go (sort by file size, group by priority tiers)
- [x] T028 [P] [US2] Write path suggestion tests in website/tools/i18n-audit/reporter_test.go (correct i18n path format for each locale)

### Implementation

- [x] T029 [US2] Implement PrioritizeMissingFiles() in website/tools/i18n-audit/analyzer.go (sort by size, categorize as P1/P2/P3)
- [x] T030 [US2] Implement GenerateMissingReport() in website/tools/i18n-audit/reporter.go (list with suggested paths)
- [x] T031 [US2] Implement list-missing command in website/tools/i18n-audit/cmd_list_missing.go (require --locale flag)
- [x] T032 [US2] Add --priority flag to list-missing command in website/tools/i18n-audit/cmd_list_missing.go

### Integration

- [x] T033 [US2] Test list-missing against real website/docs/ directory
- [x] T034 [US2] Verify priority sorting (agent-infrastructure.md and webhooks.md should be P1)

---

## Phase 5: User Story 3 - Maintain Translation Consistency (P3)

**Goal**: Detect outdated translations when source files change

**Independent Test**: Modify a source file's timestamp, run `go run . sync-check`, verify it flags the corresponding translations as outdated.

### Tests

- [x] T035 [P] [US3] Write staleness detection tests in website/tools/i18n-audit/analyzer_test.go (compare modification times, git log integration)
- [x] T036 [P] [US3] Write sync-check report tests in website/tools/i18n-audit/reporter_test.go (outdated file list, date formatting)

### Implementation

- [x] T037 [US3] Implement DetectOutdatedTranslations() in website/tools/i18n-audit/analyzer.go (compare source vs translation mtime)
- [x] T038 [US3] Implement GetGitModTime() helper in website/tools/i18n-audit/git.go (exec git log -1 --format=%ct)
- [x] T039 [US3] Implement GenerateSyncReport() in website/tools/i18n-audit/reporter.go (table with source date vs translation date)
- [x] T040 [US3] Implement sync-check command in website/tools/i18n-audit/cmd_sync_check.go

### Integration

- [x] T041 [US3] Test sync-check with artificially outdated translations (touch source file)

---

## Phase 6: Polish & Cross-Cutting Concerns

**Goal**: Production-ready features and documentation

- [ ] T042 [P] Add --format flag (table|json|markdown) to audit command in website/tools/i18n-audit/cmd_audit.go
- [ ] T043 [P] Implement GenerateJSONReport() in website/tools/i18n-audit/reporter.go
- [ ] T044 [P] Implement GenerateMarkdownReport() in website/tools/i18n-audit/reporter.go
- [ ] T045 [P] Add --output flag to write report to file in website/tools/i18n-audit/cmd_audit.go
- [ ] T046 [P] Add --min-coverage flag with exit code 1 if below threshold in website/tools/i18n-audit/cmd_audit.go
- [ ] T047 [P] Add --verbose flag for detailed file-by-file output in website/tools/i18n-audit/cmd_audit.go
- [ ] T048 [P] Implement version command in website/tools/i18n-audit/cmd_version.go
- [ ] T049 [P] Add global flags (--help, --quiet, --color, --no-color) in website/tools/i18n-audit/main.go
- [ ] T050 [P] Load .i18n-audit.yaml config file if present in website/tools/i18n-audit/config.go
- [ ] T051 [P] Implement exclusion rule matching in website/tools/i18n-audit/analyzer.go
- [ ] T052 Create README.md with installation and usage examples in website/tools/i18n-audit/README.md
- [ ] T053 Create CI/CD integration examples in website/tools/i18n-audit/examples/github-actions.yml
- [ ] T054 Create pre-commit hook example in website/tools/i18n-audit/examples/pre-commit
- [ ] T055 Run `go test ./... -race -cover` and verify >80% coverage
- [ ] T056 Build binary with `go build -o i18n-audit` and test installation
- [ ] T057 Update website documentation with i18n audit tool usage

---

## Dependencies & Execution Order

### Critical Path (Sequential)
```
Setup (T001-T005)
  ↓
Foundational (T006-T012)
  ↓
US1 Tests (T013-T015) → US1 Implementation (T016-T023) → US1 Integration (T024-T026)
  ↓
US2 Tests (T027-T028) → US2 Implementation (T029-T032) → US2 Integration (T033-T034)
  ↓
US3 Tests (T035-T036) → US3 Implementation (T037-T040) → US3 Integration (T041)
  ↓
Polish (T042-T057)
```

### Parallel Execution Opportunities

**Phase 2 (Foundational)**: T006, T007, T008, T009, T010 can run in parallel (different structs in types.go)

**Phase 3 (US1 Tests)**: T013, T014, T015 can run in parallel (different test files)

**Phase 4 (US2 Tests)**: T027, T028 can run in parallel (different test files)

**Phase 5 (US3 Tests)**: T035, T036 can run in parallel (different test files)

**Phase 6 (Polish)**: T042-T051 can run in parallel (different features, no shared state)

---

## Test Coverage Requirements

- **scanner.go**: >80% (file system operations, error handling)
- **analyzer.go**: >80% (coverage calculation, missing detection, staleness)
- **reporter.go**: >80% (all output formats: table, JSON, markdown)
- **cmd_*.go**: >60% (CLI integration, flag parsing)
- **Overall**: >75%

---

## Validation Checklist

After completing all tasks, verify:

- [ ] All tests pass: `go test ./... -race`
- [ ] Coverage meets requirements: `go test ./... -cover`
- [ ] Binary builds successfully: `go build -o i18n-audit`
- [ ] Audit command works on real website: `./i18n-audit --docs-path ../../../website/docs --i18n-path ../../../website/i18n`
- [ ] Coverage calculation matches manual count (zh-Hans: 9/16 = 56.3%)
- [ ] Missing files list is accurate (7 files: agent-infrastructure.md, bot.md, compression.md, health-monitoring.md, load-balancing.md, middleware.md, usage-tracking.md, webhooks.md)
- [ ] JSON output is valid JSON
- [ ] Markdown output renders correctly
- [ ] --min-coverage flag exits with code 1 when below threshold
- [ ] sync-check detects outdated translations
- [ ] list-missing shows correct priority order
- [ ] All CLI flags work as documented in contracts/cli-commands.md
- [ ] README.md has clear installation and usage instructions
- [ ] CI/CD examples are functional

---

## Notes

- **TDD Approach**: Write tests before implementation for all core logic (scanner, analyzer, reporter)
- **Table-Driven Tests**: Use Go idiom for test cases (see existing GoZen tests for examples)
- **Test Fixtures**: Use testdata/ directory with sample docs structure (avoid mocking file system)
- **Git Integration**: Use `git log` command for modification times (more reliable than file mtime)
- **Performance**: Target <2 seconds for audit of 100 files (use concurrent file scanning if needed)
- **Error Handling**: Graceful degradation (missing git → fallback to file mtime, missing locale → skip with warning)
- **Future Integration**: Design allows later integration as `zen docs audit` subcommand (move to internal/docsaudit/)
