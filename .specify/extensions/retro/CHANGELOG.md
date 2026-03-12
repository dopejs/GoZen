# Changelog

All notable changes to the Spec Retrospective & Self-Improvement Extension will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.2] - 2026-03-10

### Added
- GitHub Action for automated release packaging

## [1.0.1] - 2026-03-10

### Fixed
- Command naming to follow `speckit.{extension}.{command}` pattern
- Command is now `speckit.retro.analyze` with alias `speckit.retro`

## [1.0.0] - 2026-03-10

### Added
- Initial release of spec retrospective extension
- Automatic clustering for 5+ specs by component and feature
- Pattern extraction across multiple dimensions:
  - Shared constraints (constitution candidates)
  - Template gaps (missing sections)
  - Quality gate patterns (checklist items)
  - Constitution violations (principle refinements)
  - Implementation deviations (process improvements)
- Evidence-based proposals with spec citations
- Multi-group analysis with separate proposals per group
- User approval workflow for applying changes
- Optional spec archival to reduce token usage
- Constitution versioning support (MAJOR/MINOR/PATCH)
- Progressive loading for token efficiency
