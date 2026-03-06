# Feature Specification: Website Documentation Internationalization

**Feature Branch**: `012-website-i18n-docs`
**Created**: 2026-03-06
**Status**: Draft
**Input**: User description: "官网有很多文档没有实现国际化"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Identify Missing Translations (Priority: P1)

As a documentation maintainer, I need to identify which documentation pages are missing translations so that I can prioritize translation work.

**Why this priority**: This is the foundation - we need to know what's missing before we can fix it. Without this audit, we're working blind.

**Independent Test**: Run an audit script that scans the website documentation directory and generates a report showing which pages exist in English but are missing in other target languages (Chinese, Spanish, etc.). The report should be actionable and show completion percentage per language.

**Acceptance Scenarios**:

1. **Given** the website has documentation in multiple languages, **When** I run the audit tool, **Then** I see a list of all pages with their translation status per language
2. **Given** a page exists in English but not in Chinese, **When** I view the audit report, **Then** that page is marked as "missing" for Chinese
3. **Given** all pages are translated, **When** I run the audit, **Then** the report shows 100% completion for all languages

---

### User Story 2 - Translate Missing Documentation (Priority: P2)

As a documentation maintainer, I need to create translations for missing documentation pages so that users can access content in their preferred language.

**Why this priority**: Once we know what's missing, we need to actually create the translations. This delivers direct value to international users.

**Independent Test**: Take one untranslated page, create its translation following the project's i18n structure, verify it appears on the website in the correct language when that locale is selected.

**Acceptance Scenarios**:

1. **Given** a documentation page exists only in English, **When** I create a translated version following the naming convention, **Then** the translated page is accessible via the website's language selector
2. **Given** I'm translating a page, **When** I use the existing English content as reference, **Then** I maintain the same structure, headings, and code examples
3. **Given** a translated page is created, **When** I run the audit tool again, **Then** that page is no longer marked as missing

---

### User Story 3 - Maintain Translation Consistency (Priority: P3)

As a documentation maintainer, I need to ensure that when English documentation is updated, I'm aware which translations need updating so that all language versions stay in sync.

**Why this priority**: This prevents translations from becoming stale over time. It's lower priority because it's about maintenance rather than initial delivery.

**Independent Test**: Modify an English documentation page that has translations, run a sync-check tool that detects the English version has changed, verify it flags the corresponding translations as "potentially outdated".

**Acceptance Scenarios**:

1. **Given** an English page has been modified, **When** I run the sync-check tool, **Then** it identifies which translated versions may need updates
2. **Given** a translated page is updated to match the latest English version, **When** I mark it as synced, **Then** it no longer appears in the outdated list
3. **Given** multiple pages have been updated, **When** I view the sync report, **Then** I see a prioritized list based on page importance and change magnitude

---

### Edge Cases

- What happens when a new language is added to the website? (The audit should automatically detect it and show 0% completion)
- How do we handle pages that are intentionally not translated? (Need a mechanism to mark pages as "translation not required")
- What if a page is renamed or moved? (Sync-check should detect broken links between language versions)
- How do we handle partial translations? (A page might be 50% translated - should be marked as "in progress")

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST scan the website documentation directory structure and identify all documentation pages
- **FR-002**: System MUST detect which languages are currently supported by the website
- **FR-003**: System MUST generate a report showing translation completion status for each language
- **FR-004**: System MUST identify specific pages that are missing translations for each language
- **FR-005**: System MUST follow the website's existing i18n file structure and naming conventions
- **FR-006**: System MUST detect when English source pages have been modified since their translations were last updated
- **FR-007**: System MUST provide a way to mark pages as "translation not required" to exclude them from reports
- **FR-008**: System MUST generate actionable output (e.g., list of files to create, with suggested paths)

### Key Entities

- **Documentation Page**: A single markdown or MDX file containing documentation content, identified by its file path and language code
- **Language**: A supported locale (e.g., en, zh-CN, zh-TW, es) with its completion percentage
- **Translation Status**: The state of a page's translation (exists, missing, outdated, not-required)
- **Audit Report**: A summary showing translation coverage across all languages and pages

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Documentation maintainers can generate a complete translation audit report in under 30 seconds
- **SC-002**: The audit report accurately identifies 100% of missing translations with zero false positives
- **SC-003**: Translation completion percentage increases from current baseline to 90%+ for all supported languages
- **SC-004**: Time to identify which pages need translation reduces from manual review (hours) to automated report (seconds)
- **SC-005**: All newly created translations follow the correct file structure and are immediately accessible on the website

## Assumptions

- The website uses a standard i18n structure (e.g., Docusaurus i18n, Next.js i18n, or similar)
- Documentation files are in markdown or MDX format
- Language codes follow standard conventions (en, zh-CN, zh-TW, es, etc.)
- The website already has language switching functionality implemented
- Translation work will be done manually by humans (not automated translation)
- The audit tool will be run locally by maintainers, not as part of CI/CD (though it could be added later)

## Out of Scope

- Automated translation using AI/ML (translations will be done manually)
- Building the website's language switching UI (already exists)
- Translating non-documentation content (e.g., UI strings, marketing pages)
- Setting up the website's i18n infrastructure (already exists)
- Translation memory or CAT (Computer-Assisted Translation) tools
- Workflow management for translation assignments (who translates what)
