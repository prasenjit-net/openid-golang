# CI/CD Documentation

This document describes the Continuous Integration and Continuous Deployment workflows for the OpenID Connect Server.

## Overview

The project uses GitHub Actions for automated testing, building, and releasing. There are two main workflows:

1. **CI Workflow** - Automated testing and building on every push and pull request
2. **Release Workflow** - Manual release process with version bumping and multi-platform builds

## CI Workflow

**File:** `.github/workflows/ci.yml`

### Triggers

The CI workflow runs automatically on:
- Push to `main` or `develop` branches
- Pull requests targeting `main` or `develop` branches

### Jobs

#### 1. Test Backend (`test-backend`)

Tests the Go backend code:
- Sets up Go 1.21
- Caches Go modules for faster builds
- Downloads dependencies
- Runs tests with race detection and coverage
- Uploads coverage to Codecov
- Runs `go vet` for static analysis
- Checks code formatting with `gofmt`

**Commands:**
```bash
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
go vet ./...
gofmt -s -l .
```

#### 2. Test Frontend (`test-frontend`)

Tests the React admin UI:
- Sets up Node.js 20
- Caches npm dependencies
- Installs dependencies with `npm ci`
- Runs linter (optional)
- Builds the UI
- Uploads build artifacts

**Commands:**
```bash
npm ci
npm run lint
npm run build
```

#### 3. Build (`build`)

Builds binaries for multiple platforms:
- Runs after backend and frontend tests pass
- Builds on Ubuntu, macOS, and Windows
- Creates platform-specific binaries
- Uploads artifacts for download

**Platforms:**
- Linux (ubuntu-latest)
- macOS (macos-latest)
- Windows (windows-latest)

#### 4. Lint (`lint`)

Runs comprehensive linting:
- Uses golangci-lint for Go code
- Configuration in `.golangci.yml`
- Checks for common issues and code quality

#### 5. Security (`security`)

Runs security scans:
- Uses Gosec security scanner
- Uploads results in SARIF format
- Integrates with GitHub Security tab

### Artifacts

Build artifacts are retained for 7 days:
- `ui-build` - React production build
- `openid-server-{os}` - Platform-specific binaries

## Release Workflow

**File:** `.github/workflows/release.yml`

### Triggers

The release workflow is **manually triggered** via GitHub Actions UI:

1. Go to **Actions** tab in GitHub
2. Select **Release** workflow
3. Click **Run workflow**
4. Choose parameters:
   - **Version bump type**: patch/minor/major
   - **Pre-release**: true/false

### Version Bumping

The workflow uses semantic versioning (MAJOR.MINOR.PATCH):

- **Patch** (0.1.0 → 0.1.1): Bug fixes and minor changes
- **Minor** (0.1.0 → 0.2.0): New features, backward compatible
- **Major** (0.1.0 → 1.0.0): Breaking changes

### Release Process

1. **Version Calculation**
   - Reads current version from `VERSION` file
   - Calculates new version based on bump type
   - Updates `VERSION` file

2. **Git Operations**
   - Commits version change
   - Creates and pushes git tag (e.g., `v0.1.1`)
   - Pushes changes to main branch

3. **Build Process**
   - Builds React admin UI
   - Compiles Go binaries for 5 platforms:
     - Linux AMD64
     - Linux ARM64
     - macOS AMD64 (Intel)
     - macOS ARM64 (Apple Silicon)
     - Windows AMD64
   - Embeds version information in binaries
   - Creates SHA256 checksums

4. **Release Creation**
   - Generates changelog from git commits
   - Creates GitHub release
   - Uploads all binaries and checksums
   - Adds installation instructions

### Build Artifacts

Released binaries include:
- `openid-server-linux-amd64`
- `openid-server-linux-arm64`
- `openid-server-darwin-amd64`
- `openid-server-darwin-arm64`
- `openid-server-windows-amd64.exe`
- `checksums.txt` - SHA256 checksums

### Version Information

Version is embedded in the binary at build time:

```bash
go build -ldflags="-X main.Version=0.1.1" ./cmd/server
```

Check version:
```bash
./openid-server --version
# Output: OpenID Connect Server v0.1.1
```

## Configuration Files

### `.golangci.yml`

Linter configuration with enabled rules:
- Code quality checks
- Style enforcement
- Security scanning
- Performance optimization detection
- Import organization

### `VERSION`

Simple text file containing current version:
```
0.1.0
```

## Local Testing

### Test CI Locally

**Backend tests:**
```bash
go test -v -race -coverprofile=coverage.txt ./...
go vet ./...
gofmt -s -l .
```

**Frontend:**
```bash
cd ui/admin
npm ci
npm run build
```

**Linting:**
```bash
golangci-lint run --timeout=5m
```

### Simulate Release Build

```bash
# Build UI
cd ui/admin && npm run build && cd ../..

# Copy to embed location
mkdir -p internal/ui/admin
cp -r ui/admin/dist internal/ui/admin/

# Build with version
go build -ldflags="-X main.Version=dev" -o bin/openid-server ./cmd/server

# Test version
./bin/openid-server --version
```

## Troubleshooting

### Common Issues

**1. Tests failing in CI but passing locally**
- Ensure all dependencies are committed
- Check for environment-specific issues
- Verify test fixtures are portable

**2. Build failing on specific platform**
- Check platform-specific code
- Verify file paths use `filepath.Join()`
- Test cross-compilation locally

**3. Release workflow fails**
- Check `VERSION` file exists and is valid
- Ensure GitHub token has write permissions
- Verify git config is correct

### Manual Release Recovery

If release workflow fails partway:

```bash
# Check current state
git tag -l
cat VERSION

# If tag was created but release failed
git push origin :refs/tags/v0.1.1  # Delete remote tag
git tag -d v0.1.1                   # Delete local tag

# If version was bumped
git reset --hard HEAD~1             # Reset to previous commit
git push origin main --force        # Force push (use with caution)

# Re-run workflow
```

## Best Practices

### Before Committing

1. Run tests locally: `go test ./...`
2. Run linter: `golangci-lint run`
3. Format code: `gofmt -w .`
4. Build project: `./build.sh`

### Pull Requests

- All CI checks must pass
- Maintain test coverage
- Update documentation if needed
- Follow commit message conventions

### Releases

- Use semantic versioning appropriately
- Write clear release notes
- Test release candidate before marking as stable
- Mark experimental features as pre-release

## Monitoring

### GitHub Actions

Monitor workflow runs:
- Go to **Actions** tab
- View workflow history
- Download artifacts
- Check logs for failures

### Badges

Add to README.md:
```markdown
[![CI](https://github.com/prasenjit-net/openid-golang/actions/workflows/ci.yml/badge.svg)](https://github.com/prasenjit-net/openid-golang/actions/workflows/ci.yml)
```

## Security

### Secrets Management

Required secrets:
- `GITHUB_TOKEN` - Automatically provided by GitHub
- Optional: `CODECOV_TOKEN` for coverage reports

### Security Scanning

The workflow includes:
- Gosec for Go security issues
- Dependency vulnerability scanning
- SARIF upload to GitHub Security

## Future Improvements

Potential enhancements:
- [ ] Add E2E tests
- [ ] Docker image building and publishing
- [ ] Deploy to staging environment
- [ ] Performance benchmarking
- [ ] Automated security updates
- [ ] Release notes generation from PRs
- [ ] Slack/Discord notifications

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Semantic Versioning](https://semver.org/)
- [golangci-lint](https://golangci-lint.run/)
- [Gosec](https://github.com/securego/gosec)
