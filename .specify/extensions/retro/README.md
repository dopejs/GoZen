# Spec Retrospective & Self-Improvement Extension

A spec-kit extension that enables continuous improvement by analyzing completed specs to extract shared constraints, patterns, and lessons learned, then proposing improvements to your project's constitution, templates, and checklists.

## Overview

As your project accumulates specs, this extension helps you:

- **Extract cross-cutting patterns** from completed implementations
- **Identify shared constraints** that should be codified in your constitution
- **Discover template gaps** where manual additions are repeatedly needed
- **Build quality checklists** based on recurring validation needs
- **Reduce token usage** by archiving analyzed specs after extracting their lessons

## Philosophy

**Self-Improvement Loop**: Each completed spec is a learning opportunity. The retro process distills implementation experience into reusable governance rules, better templates, and quality gates that make future specs higher quality with less ambiguity.

**Constitution-First**: The constitution is the only artifact loaded by all future specs. Extracting shared constraints into the constitution has maximum leverage—it improves every subsequent spec automatically.

**No New Artifacts**: This extension does NOT create summary.md or other new file types. It strengthens existing infrastructure (constitution, templates, checklists) that spec-kit commands already use.

## Installation

### From GitHub (recommended)

```bash
specify extension add --from https://github.com/dopejs/speckit-retro-extension
```

### From Local Path (for development)

```bash
specify extension add --from /path/to/speckit-retro-extension
```

## Usage

### Basic Usage

Analyze a range of completed specs:

```bash
/speckit.retro 015-019
```

Analyze specific specs:

```bash
/speckit.retro 015,017,019
```

Analyze all specs:

```bash
/speckit.retro all
```

### Workflow

1. **Identify specs** — Specify which completed specs to analyze
2. **Automatic clustering** — For 5+ specs, automatically groups by topic/component
3. **Select groups** — Choose which groups to analyze (or analyze all)
4. **Pattern extraction** — Identifies cross-spec patterns in:
   - Shared constraints (constitution candidates)
   - Template gaps (missing sections)
   - Quality gates (checklist items)
   - Constitution violations (principle refinements)
   - Implementation deviations (process improvements)
5. **Review proposals** — Structured improvement proposals with evidence
6. **Apply changes** — Select which improvements to apply
7. **Optional archival** — Move analyzed specs to `.archive/` to reduce token usage

### Example Output

```markdown
## Group 1: Backend & API - Improvement Proposals

### Constitution Amendments

#### Amendment 1.1: API Response Time Limits

**Type**: New constraint
**Rationale**: Specs 017, 019 both implemented 10-second timeout mechanisms
**Proposed Text**: All API endpoints MUST respond within 10 seconds...
**Version Bump**: MINOR

### Template Updates

#### Template Update 1.1: Add Error Handling Strategy

**Template**: plan-template.md
**Rationale**: Specs 017, 019, 020 all added this section manually
**Proposed Diff**: [shows exact changes]

### Checklist Additions

#### Checklist Addition 1.1: API Design Checklist

**Items**:
- [ ] CHK-API-001: All endpoints have timeout handling
- [ ] CHK-API-002: Error responses follow consistent format
...
```

## Features

### Automatic Clustering (5+ specs)

When analyzing many specs, the extension automatically clusters them by:

**Component-based**: Backend/API, Frontend/UI, CLI/Tooling, Infrastructure, Testing
**Feature-based**: Stability, Performance, Security, Data, User Experience

This ensures proposals are focused and domain-specific rather than mixing unrelated constraints.

### Evidence-Based Proposals

Every proposal cites specific specs as evidence. Minimum 2-3 specs must show the same pattern before it's proposed as a shared constraint.

### Safe Application

- User approval required for all changes
- Constitution versioning follows governance rules (MAJOR/MINOR/PATCH)
- All changes are git-tracked and reversible
- Original specs preserved in `.archive/` if archived

### Token Efficiency

- Progressive loading (specs loaded incrementally)
- Minimal context extraction (only relevant sections)
- Optional archival reduces future token usage by 90%+

## When to Run Retro

**Recommended cadence**: After completing every 5-10 specs

**Good times to run**:
- After a major feature milestone
- Before starting a new development phase
- When you notice repeated patterns across recent specs
- When constitution feels outdated or incomplete

**Signs you need retro**:
- Specs repeatedly add the same manual sections
- Similar quality issues appear across multiple specs
- Constitution violations have similar justifications
- Token usage is growing due to many historical specs

## Requirements

- spec-kit >= 0.1.0
- Existing spec-kit project with completed specs
- Constitution file at `.specify/memory/constitution.md`

## Configuration

No configuration needed. The extension works with your existing spec-kit setup.

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Test on real projects
4. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

- Issues: https://github.com/dopejs/speckit-retro-extension/issues
- Discussions: https://github.com/dopejs/speckit-retro-extension/discussions

## Related

- [spec-kit](https://github.com/github/spec-kit) - The core spec-kit framework
- [Extension Development Guide](https://github.com/github/spec-kit/blob/main/extensions/EXTENSION-DEVELOPMENT-GUIDE.md)
