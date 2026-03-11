# speckit-retro-extension

Spec-kit extension for retrospective analysis and continuous self-improvement.

## Quick Links

- [Installation & Usage](README.md)
- [Changelog](CHANGELOG.md)
- [License](LICENSE)

## Structure

```
speckit-retro-extension/
├── extension.yml          # Extension manifest
├── README.md             # Documentation
├── LICENSE               # MIT License
├── CHANGELOG.md          # Version history
└── commands/
    └── retro.md          # /speckit.retro command implementation
```

## Local Testing

To test this extension locally before publishing:

```bash
cd ~/Code/speckit-retro-extension
specify extension add --from .
```

Then in your spec-kit project:

```bash
/speckit.retro 015-019
```

## Publishing

1. Update version in `extension.yml`
2. Update `CHANGELOG.md`
3. Commit changes
4. Create git tag: `git tag v1.0.0`
5. Push tag: `git push origin v1.0.0`
6. Create GitHub release from tag

Users can then install via:

```bash
specify extension add --from https://github.com/dopejs/speckit-retro-extension
```
