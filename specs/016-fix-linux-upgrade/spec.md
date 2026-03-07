# Feature Specification: Fix Linux Upgrade Failure

**Feature Branch**: `016-fix-linux-upgrade`
**Created**: 2026-03-07
**Status**: Draft
**Input**: User description: "zen upgrade fails on Linux (CentOS/Debian) with 'install failed: exit status 1' when replacing the running binary, even as root user"

## Context

On Linux systems, running `zen upgrade <version>` fails after successfully downloading the new binary. The error occurs at the installation step when the tool attempts to overwrite the currently-running `zen` binary. The root cause is ETXTBSY ("text file busy") — Linux prevents writing to an executable that is currently running. This affects all Linux distributions (confirmed on CentOS and Debian) regardless of user privilege level (the user is root, so it is not a permissions issue). macOS does not exhibit this behavior.

The current installation logic uses `os.OpenFile(dst, O_WRONLY|O_TRUNC)` to overwrite in place, which triggers ETXTBSY. The sudo fallback (`sudo cp`) hits the same error because `cp` also opens the destination for writing.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Upgrade on Linux Succeeds (Priority: P1)

A user running `zen upgrade` on a Linux system (as root or non-root) can successfully upgrade the binary without encountering "install failed" errors.

**Why this priority**: This is a blocking bug — Linux users cannot upgrade at all. It must work before v3.0.0 release.

**Independent Test**: Run `zen upgrade <version>` on a Linux system where `zen` is currently running. The upgrade completes successfully and the new version is active.

**Acceptance Scenarios**:

1. **Given** a Linux system with `zen` installed at `/usr/local/bin/zen`, **When** the user runs `zen upgrade 3.0.0-alpha.21`, **Then** the download completes, the binary is replaced, and the success message is printed.
2. **Given** a Linux system where the user is root, **When** running `zen upgrade`, **Then** the upgrade does not attempt unnecessary `sudo` and completes directly.
3. **Given** a Linux system where `zen` is running (the upgrade command itself), **When** the binary replacement occurs, **Then** the old binary is removed-then-replaced (not overwritten in place), avoiding ETXTBSY.

---

### User Story 2 - Upgrade on macOS Continues to Work (Priority: P1)

The fix for Linux must not regress the existing macOS upgrade path, including ad-hoc codesigning.

**Why this priority**: macOS is the primary user base; regression would be worse than the Linux bug.

**Independent Test**: Run `zen upgrade <version>` on macOS. The upgrade completes successfully with codesign applied.

**Acceptance Scenarios**:

1. **Given** a macOS system with `zen` installed, **When** the user runs `zen upgrade <version>`, **Then** the binary is replaced and codesigned as before.
2. **Given** a macOS system where direct file operations fail (permission denied), **When** sudo fallback is used, **Then** codesign also runs with sudo.

---

### User Story 3 - Graceful Error Messages (Priority: P2)

When installation fails for an unexpected reason (disk full, read-only filesystem), the error message should be clear and actionable.

**Why this priority**: Good error messages help users self-diagnose, reducing support burden.

**Independent Test**: Simulate a write failure (e.g., read-only destination) and verify the error message includes the reason and suggests manual steps.

**Acceptance Scenarios**:

1. **Given** a system where the destination directory is read-only, **When** the upgrade fails, **Then** the error message includes the underlying OS error and a suggestion (e.g., "try running with sudo" or "check filesystem permissions").
2. **Given** an upgrade failure after download succeeds, **When** the error is displayed, **Then** the downloaded file path is mentioned so the user can install manually.

---

### Edge Cases

- What happens if the destination binary is a symlink? The replacement should follow the symlink and replace the target.
- What happens if another process has the binary open (e.g., a running daemon)? On Linux, `os.Remove` on a running binary succeeds (the inode persists until the process exits), so this is safe.
- What happens on a read-only filesystem? The error should be caught and reported clearly.
- What if the temp file and the destination are on different filesystems? `os.Rename` will fail; the code must handle cross-device fallback.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The upgrade command MUST replace the binary using a remove-then-write strategy (not overwrite-in-place) to avoid ETXTBSY on Linux.
- **FR-002**: The upgrade command MUST resolve symlinks before replacing, so that the actual binary file is replaced, not the symlink.
- **FR-003**: When the user is already root, the upgrade command MUST NOT attempt `sudo` (it is redundant and may fail if sudo is not installed).
- **FR-004**: When a direct file operation fails due to permissions (not ETXTBSY), the upgrade command MUST fall back to sudo for the remove-and-copy operation.
- **FR-005**: The upgrade command MUST preserve the existing file permissions of the replaced binary (or default to 0755 if creating new).
- **FR-006**: The sudo fallback MUST also use the remove-then-write strategy (not `sudo cp` which opens for writing).
- **FR-007**: The upgrade MUST continue to apply ad-hoc codesigning on macOS after binary replacement.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `zen upgrade` completes successfully on Linux (CentOS 7+, Debian 10+, Ubuntu 20.04+) when run as root.
- **SC-002**: `zen upgrade` completes successfully on Linux when run as a non-root user with write permissions to the binary location.
- **SC-003**: `zen upgrade` continues to work on macOS without regression (codesigning applied).
- **SC-004**: All existing upgrade tests pass, plus new tests covering the Linux-specific code path.
- **SC-005**: Error messages for failed installations include the underlying OS error description.

## Assumptions

- The ETXTBSY error is the root cause (confirmed by user report: download succeeds, install fails, user is root).
- The `os.Remove()` + create-new-file approach is safe on Linux for running binaries (the running process keeps its inode reference until exit).
- Cross-device rename (temp file on different filesystem than target) will be handled by falling back to copy.
