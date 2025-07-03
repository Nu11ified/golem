# 🚀 Automated Release System Guide

This repository uses GitHub Actions to automatically build, test, and release multi-platform binaries. Here's how it works:

## 🔄 Workflow Overview

We have three main workflows:

1. **CI Workflow** (`ci.yml`) - Runs on every push and PR
2. **Auto Release** (`auto-release.yml`) - Detects release-worthy commits
3. **Release Workflow** (`release.yml`) - Builds and publishes releases

## 📦 Automatic Release Triggers

### Method 1: Commit Message Keywords

The system automatically detects releases based on commit messages. Use these patterns:

```bash
# For patch releases (bug fixes)
git commit -m "fix: resolve authentication issue"
git commit -m "fix(auth): handle edge case in token validation"

# For minor releases (new features)
git commit -m "feat: add new component system"
git commit -m "feat(cli): add interactive project setup"

# For major releases (breaking changes)
git commit -m "feat!: redesign API structure"
git commit -m "breaking: remove deprecated methods"

# Explicit release commits
git commit -m "release: v1.2.3"
git commit -m "Release v1.2.3"
git commit -m "[release] New version with bug fixes"
```

### Method 2: Manual Release

You can trigger releases manually:

1. **Via GitHub UI**: Go to Actions → Release → Run workflow
2. **Via Git Tags**: 
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

### Method 3: Push to Main (Auto-increment)

When you push to main without release keywords, it auto-increments patch version if there are new commits since the last tag.

## 🏗️ Build Matrix

Each release builds for these platforms:

- **Linux**: x64, ARM64
- **macOS**: Intel (x64), Apple Silicon (ARM64)
- **Windows**: x64

## 📋 Version Bumping Rules

The system follows semantic versioning:

| Commit Type | Version Bump | Example |
|-------------|-------------|---------|
| `fix:` | Patch | 1.0.0 → 1.0.1 |
| `feat:` | Minor | 1.0.0 → 1.1.0 |
| `feat!:` or `breaking:` | Major | 1.0.0 → 2.0.0 |
| Manual | User-defined | Any version |

## 🔧 Workflow Files

### CI Workflow (`ci.yml`)
- Runs on every push/PR to main/develop
- Tests on Ubuntu, Windows, macOS
- Runs linting and builds
- Provides feedback on code quality

### Auto Release (`auto-release.yml`)
- Monitors commits to main branch
- Analyzes commit messages for release patterns
- Calculates new version numbers
- Triggers release workflow when needed
- Ignores documentation-only changes

### Release Workflow (`release.yml`)
- Builds binaries for all platforms
- Creates compressed archives
- Generates changelog from commits
- Uploads to GitHub Releases
- Supports multiple trigger methods

## 🎯 Release Assets

Each release includes:

```
golem-v1.0.0-linux-amd64.tar.gz
golem-v1.0.0-linux-arm64.tar.gz
golem-v1.0.0-darwin-amd64.tar.gz
golem-v1.0.0-darwin-arm64.tar.gz
golem-v1.0.0-windows-amd64.zip
```

## 🛠️ Usage Examples

### Regular Development
```bash
# Regular commits (no release)
git commit -m "docs: update README"
git commit -m "refactor: improve code organization"
git push origin main  # → No release triggered

# Bug fix (patch release)
git commit -m "fix: resolve memory leak in parser"
git push origin main  # → Triggers v1.0.1 release

# New feature (minor release)
git commit -m "feat: add dark mode support"
git push origin main  # → Triggers v1.1.0 release
```

### Manual Release
```bash
# Create specific version
git tag v2.0.0
git push origin v2.0.0  # → Triggers v2.0.0 release

# Or use GitHub Actions UI
# Go to Actions → Release → Run workflow
# Enter version: v2.0.0
```

### Breaking Changes
```bash
# Major version bump
git commit -m "feat!: redesign CLI interface"
git push origin main  # → Triggers v2.0.0 release

# Or
git commit -m "breaking: remove deprecated API"
git push origin main  # → Triggers v2.0.0 release
```

## 🔍 Monitoring Releases

1. **GitHub Actions**: Monitor workflow runs in the Actions tab
2. **Releases Page**: Check the releases section for published versions
3. **Commit History**: View which commits triggered releases

## 🐛 Troubleshooting

### Release Not Triggered
- Check commit message format
- Ensure you're pushing to the `main` branch
- Verify the commit contains actual code changes (not just docs)

### Build Failures
- Check the Actions tab for detailed error logs
- Ensure all tests pass locally: `go test ./...`
- Verify dependencies are up to date: `go mod tidy`

### Version Conflicts
- Check existing tags: `git tag -l`
- Ensure new version is higher than the latest tag
- Use semantic versioning format (v1.2.3)

## 🔒 Security & Permissions

The workflows use `GITHUB_TOKEN` which has limited permissions:
- Read repository contents
- Write releases and tags
- No access to secrets or external services

## 📊 Metrics

Track release effectiveness:
- **Frequency**: How often releases are created
- **Success Rate**: Percentage of successful builds
- **Download Stats**: View release download counts
- **Build Time**: Monitor performance improvements

## 🤝 Contributing

When contributing:
1. Use conventional commit messages
2. Test locally before pushing
3. Check CI status before merging PRs
4. Use descriptive commit messages for better changelogs

---

*This automated system ensures consistent, reliable releases while minimizing manual overhead.* 