---
description: Perform retrospective analysis on completed specs to extract shared constraints and improve constitution, templates, and checklists through self-improvement.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Goal

Analyze completed specs to identify cross-cutting patterns, constraints, and lessons learned, then propose improvements to the project's constitution, templates, and checklists. This enables speckit to continuously improve itself based on real implementation experience.

## Operating Philosophy

**Self-Improvement Loop**: Each completed spec is a learning opportunity. The retro process distills implementation experience into reusable governance rules, better templates, and quality gates that make future specs higher quality with less ambiguity.

**Constitution-First**: The constitution is the only artifact loaded by all future specs. Extracting shared constraints into the constitution has maximum leverage—it improves every subsequent spec automatically.

**No New Artifacts**: This command does NOT create summary.md or other new file types. It strengthens existing infrastructure (constitution, templates, checklists) that speckit commands already use.

## Execution Steps

### 1. Identify Completed Specs

Ask the user which specs to analyze. Suggested formats:
- Range: "015-019" (analyze specs 015 through 019)
- List: "015,017,019" (analyze specific specs)
- All: "all" (analyze all specs in /specs/)

Parse the input and build a list of spec directories to analyze.

### 2. Cluster Specs by Topic (if analyzing 5+ specs)

**Skip this step if analyzing fewer than 5 specs.**

For large-scale analysis (5+ specs), automatically cluster specs by topic before pattern extraction to ensure high-quality, focused proposals.

#### 2.1 Load Minimal Context for Clustering

For each spec, read only:
- **spec.md**: First 50 lines (Overview/Context section)
- **plan.md**: Technology & Architecture Constraints section

Extract key indicators:
- Primary components mentioned (daemon, proxy, CLI, web UI, config, TUI, etc.)
- Technology stack (Go packages, React, SQLite, etc.)
- Feature category keywords (stability, routing, monitoring, migration, etc.)

#### 2.2 Perform Automatic Clustering

Group specs by similarity using these heuristics:

**Component-based clustering** (adapt to your project structure):
- Backend/API: specs mentioning server, API endpoints, business logic, data access
- Frontend/UI: specs mentioning UI components, user interactions, styling, client-side logic
- CLI/Tooling: specs mentioning command-line interface, scripts, automation
- Infrastructure: specs mentioning deployment, configuration, monitoring, logging
- Testing & Quality: specs mentioning test coverage, CI/CD, integration tests

**Feature-based clustering** (if component clustering produces groups >8 specs):
- Stability & Reliability: error handling, recovery, resilience, fault tolerance
- Performance & Scalability: optimization, caching, concurrency, load handling
- Security & Privacy: authentication, authorization, data protection, validation
- Data & Storage: database, schema, migration, persistence
- User Experience: usability, accessibility, responsiveness, feedback

#### 2.3 Present Clustering Results

Output clustering summary:

```markdown
## Spec Clustering Results

Analyzed 20 specs, grouped into 4 clusters:

### Group 1: Backend & API (7 specs)
- 015-user-authentication
- 017-api-rate-limiting
- 018-data-validation
- 019-caching-strategy
- 020-error-handling
- ...

**Focus areas**: API design, data processing, error handling, performance

### Group 2: Frontend & UI (5 specs)
- 003-responsive-layout
- 011-form-validation
- 016-accessibility-improvements
- ...

**Focus areas**: Component design, user interactions, styling, accessibility

### Group 3: Infrastructure & Deployment (4 specs)
- 005-logging-system
- 006-monitoring-dashboard
- 008-ci-cd-pipeline
- ...

**Focus areas**: Observability, deployment automation, configuration management

### Group 4: Testing & Quality (4 specs)
- 008-integration-tests
- ...

**Focus areas**: Test coverage, quality gates, automated testing
```

#### 2.4 Ask User to Select Groups

Present options:

```
Which groups would you like to analyze?
- [ ] Group 1: Daemon & Proxy Stability (7 specs)
- [ ] Group 2: CLI & Config Management (5 specs)
- [ ] Group 3: Web UI & Monitoring (4 specs)
- [ ] Group 4: Testing & Quality (4 specs)
- [ ] All groups (analyze each separately, generate per-group proposals)
- [ ] Skip clustering (analyze all specs together)
```

Wait for user selection before proceeding.

**If user selects multiple groups**: Analyze each group independently and generate separate proposal sections for each.

**If user selects "Skip clustering"**: Proceed with all specs in a single analysis (may produce lower-quality cross-domain proposals).

### 3. Load Spec Artifacts

For each spec in the selected group(s), load:

**From spec.md**:
- Functional requirements
- Non-functional requirements
- User stories
- Edge cases

**From plan.md**:
- Architecture decisions and rationale
- Technology choices
- Constitution Check section (violations, complexity justifications)
- Phase breakdown

**From tasks.md**:
- Task structure and organization patterns
- Dependency patterns
- Parallelization markers

**From checklists/** (if exists):
- Quality dimensions checked
- Recurring validation patterns

**From implementation** (if merged):
- Check git log for the feature branch to understand what was actually built
- Look for deviations between plan and implementation

### 3. Pattern Extraction

Analyze loaded specs across these dimensions:

#### A. Shared Constraints (Constitution Candidates)

Identify rules that appear across multiple specs:
- Technology choices that became de facto standards
- Architecture patterns repeatedly used
- Performance/security requirements that recur
- Testing strategies applied consistently
- Forbidden patterns that caused issues

**Example**: If 3+ specs all avoid nested API responses beyond 3 levels, that's a constraint worth codifying.

#### B. Template Gaps

Identify sections frequently added manually that should be in templates:
- Missing sections in spec-template.md (e.g., "Performance Considerations")
- Missing phases in plan-template.md
- Missing task categories in tasks-template.md

**Example**: If every spec adds a "Migration Strategy" section, add it to spec-template.

#### C. Quality Gate Patterns

Identify validation checks that should become default checklists:
- Security checks repeatedly needed
- Performance validation patterns
- UX quality dimensions
- API design principles

**Example**: If multiple specs check "rate limiting for batch operations", add it to a default checklist.

#### D. Constitution Violations

Review Constitution Check sections across specs:
- Which principles are frequently violated?
- Are violations justified (complexity trade-offs) or avoidable?
- Do violation patterns suggest the principle needs refinement?

**Example**: If Principle II is violated in 5 specs with similar justifications, the principle may need adjustment.

#### E. Implementation Deviations

Compare plans vs actual implementation:
- What changed during implementation and why?
- Were there recurring surprises or unknowns?
- Did certain types of tasks consistently take longer than expected?

**Example**: If integration tasks consistently reveal missing error handling, add "error handling strategy" to plan-template.

### 4. Generate Improvement Proposals

**If analyzing multiple groups**: Generate separate proposal sections for each group with clear group headers.

**If analyzing a single group or all specs together**: Generate a unified proposal.

Output a structured proposal document with three sections per group:

#### Proposed Constitution Amendments

For each proposed amendment:
- **Type**: New principle | Principle modification | New constraint
- **Rationale**: Which specs demonstrate this pattern (cite spec numbers)
- **Proposed Text**: Exact wording to add/modify
- **Impact**: Which future specs will benefit
- **Version Bump**: MAJOR | MINOR | PATCH (per constitution governance rules)

**Format** (for multi-group analysis):
```markdown
## Group 1: Daemon & Proxy Stability - Improvement Proposals

### Constitution Amendments

#### Amendment 1.1: Concurrency Limits for Proxy

**Type**: New constraint (add to "Technology & Architecture Constraints")

**Rationale**: Specs 017, 019 both implemented concurrency limiting (semaphore pattern with 100 concurrent requests). This pattern should be codified to prevent resource exhaustion.

**Proposed Text**:
> - **Proxy Concurrency**: The daemon proxy MUST enforce a configurable concurrency limit (default 100 concurrent requests) using semaphore pattern to prevent resource exhaustion under load.

**Impact**: Future proxy-related specs will include concurrency control from the planning phase.

**Version Bump**: MINOR (new constraint)

### Template Updates

#### Template Update 1.1: Add Stability Testing to plan-template.md

**Template**: `.specify/templates/plan-template.md`

**Change Type**: Add section

**Rationale**: Specs 017, 019, 020 all added "Stability Testing" sections manually for daemon-related features.

**Proposed Diff**:
```diff
+ ## Stability Testing (for daemon/proxy features)
+
+ - Load testing strategy (concurrent requests, duration)
+ - Failure injection scenarios (panic, timeout, connection loss)
+ - Resource leak detection (goroutines, connections, memory)
+ - Recovery validation (auto-restart, graceful degradation)
```

### Checklist Additions

#### Checklist Addition 1.1: Daemon Stability Checklist

**Checklist**: Create `.specify/templates/daemon-stability-checklist-template.md`

**Items**:
- [ ] CHK-DAEMON-001: Panic recovery middleware implemented
- [ ] CHK-DAEMON-002: Goroutine leak detection with baseline comparison
- [ ] CHK-DAEMON-003: Connection pool cleanup on cache invalidation
- [ ] CHK-DAEMON-004: Load test with 100 concurrent requests passes
- [ ] CHK-DAEMON-005: Auto-restart with exponential backoff tested

**Rationale**: Specs 017, 019 both needed these checks. Creating a dedicated daemon stability checklist catches these requirements during planning.

---

## Group 2: CLI & Config Management - Improvement Proposals

### Constitution Amendments

#### Amendment 2.1: Config Migration Safety

...
```

**Format** (for single-group or unified analysis):
```markdown
### Amendment 1: API Response Nesting Limit

**Type**: New constraint (add to "Technology & Architecture Constraints")

**Rationale**: Specs 015, 017, 019 all independently limited API response nesting to 3 levels for performance and client parsing simplicity. This pattern should be codified.

**Proposed Text**:
> - **API Design**: Response bodies MUST NOT nest objects deeper than 3 levels. Use flat structures with references (IDs) for deep relationships.

**Impact**: Prevents future specs from creating deeply nested APIs that cause client-side parsing issues.

**Version Bump**: MINOR (new constraint)
```

#### Proposed Template Updates

For each template update:
- **Template**: Which template file
- **Change Type**: Add section | Modify section | Remove section
- **Rationale**: Which specs needed this manually
- **Proposed Diff**: Show before/after

**Format**:
```markdown
### Template Update 1: Add Performance Considerations to plan-template.md

**Template**: `.specify/templates/plan-template.md`

**Change Type**: Add section

**Rationale**: Specs 016, 017, 018, 019 all added "Performance Considerations" sections manually. This should be a standard plan section.

**Proposed Diff**:
```diff
+ ## Performance Considerations
+
+ - Expected load characteristics
+ - Performance targets (latency, throughput)
+ - Bottleneck analysis
+ - Optimization strategy
```
```

#### Proposed Checklist Additions

For each checklist addition:
- **Checklist**: Which checklist file (or new checklist to create)
- **Items**: New checklist items to add
- **Rationale**: Which specs would have caught issues earlier

**Format**:
```markdown
### Checklist Addition 1: Rate Limiting Check

**Checklist**: `.specify/templates/checklist-template.md` (or create `api-checklist-template.md`)

**Items**:
- [ ] CHK-API-001: Batch operations have rate limiting
- [ ] CHK-API-002: Rate limit errors return 429 with Retry-After header
- [ ] CHK-API-003: Rate limits documented in API contracts

**Rationale**: Specs 016, 019 both discovered missing rate limiting during implementation. Adding this to default API checklist catches it during planning.
```

### 5. Archive Completed Specs (Optional)

After extracting improvements, optionally archive the analyzed specs:

Ask user: "Would you like to archive these specs? This will move original files to `specs/.archive/[NNN]-feature-name/` to reduce future token usage."

If yes:
- Create `specs/.archive/` directory if it doesn't exist
- For each analyzed spec:
  - Move entire spec directory to `.archive/`
  - Leave a minimal index file at original location (optional)

**Minimal index format** (if user wants it):
```markdown
# [NNN] - [Feature Name] (Archived)

Archived: [date]
Location: `specs/.archive/[NNN]-feature-name/`
Constitution updates: [list amendment numbers from retro]
```

### 6. User Review and Approval

Present the complete proposal and ask:

"Review the proposed improvements above. Which changes would you like to apply?"

Options:
- [ ] Apply all constitution amendments
- [ ] Apply all template updates
- [ ] Apply all checklist additions
- [ ] Apply selected items (specify which)
- [ ] Save proposal for later review
- [ ] Cancel (no changes)

### 7. Apply Approved Changes

For each approved change:

**Constitution amendments**:
1. Read current `.specify/memory/constitution.md`
2. Apply the amendment
3. Update version number per governance rules
4. Update Sync Impact Report (HTML comment at top)
5. Write updated constitution

**Template updates**:
1. Read the template file
2. Apply the diff
3. Write updated template

**Checklist additions**:
1. Read or create the checklist template
2. Add new items with proper CHK-### IDs
3. Write updated checklist

### 8. Generate Retro Summary

Output a concise summary:

```markdown
## Retro Summary

**Specs Analyzed**: [list with group breakdown if applicable]
**Groups**: [number of groups, or "unified analysis"]
**Patterns Identified**: [count per group if applicable]
**Changes Applied**:
- Constitution: [count] amendments (version [old] → [new])
- Templates: [count] updates
- Checklists: [count] additions

**Per-Group Breakdown** (if multi-group analysis):
- Group 1 (Daemon & Proxy): [X] amendments, [Y] template updates, [Z] checklist items
- Group 2 (CLI & Config): [X] amendments, [Y] template updates, [Z] checklist items
- ...

**Next Steps**:
- New specs will automatically benefit from updated constitution and templates
- Existing in-progress specs may want to incorporate new checklist items
- Consider running retro again after completing next 5-10 specs
```

## Operating Principles

### Context Efficiency

- **Progressive loading**: Load specs incrementally, not all at once
- **Pattern-focused**: Extract high-signal patterns, not exhaustive documentation
- **Minimal output**: Proposals should be concise and actionable

### Analysis Guidelines

- **Evidence-based**: Every proposal must cite specific specs as evidence
- **Cross-spec patterns only**: Don't propose rules based on a single spec (minimum 2-3 specs showing the same pattern)
- **Respect constitution governance**: Follow versioning and amendment rules
- **No speculation**: Only propose constraints actually demonstrated in completed specs
- **Group-focused proposals**: When analyzing multiple groups, ensure proposals are relevant to the group's domain (don't mix daemon constraints with web UI constraints)

### Safety

- **User approval required**: Never auto-apply constitution changes
- **Preserve originals**: Archive moves files, doesn't delete them
- **Reversible**: All changes are git-tracked and can be reverted

## Context

$ARGUMENTS
