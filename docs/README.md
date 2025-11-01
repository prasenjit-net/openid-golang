# Documentation for OpenID Connect Identity Server

This directory contains all documentation for the OpenID Connect Identity Server project.

## 🌐 Published Documentation

The documentation is published on GitHub Pages at:
**https://prasenjit-net.github.io/openid-golang/**

## 📚 Documentation Structure

### Getting Started
- [Getting Started Guide](GETTING_STARTED.md) - Complete setup tutorial
- [Quick Start](QUICKSTART.md) - Fast setup for experienced developers
- [Docker Quick Start](DOCKER_QUICKSTART.md) - Docker deployment
- [Setup Wizard](SETUP_WIZARD.md) - Interactive configuration

### Core Documentation
- [API Reference](API.md) - Complete API documentation
- [Architecture](ARCHITECTURE.md) - System design
- [Configuration](CONFIGURATION.md) - Configuration reference
- [Storage](STORAGE.md) - Storage backends

### Development
- [Dev Setup](DEV_SETUP.md) - Development environment
- [Testing](TESTING.md) - Testing guide
- [Contributing](CONTRIBUTING.md) - Contribution guidelines
- [Structure](STRUCTURE.md) - Project structure

### Implementation & Compliance
- [OIDC Compliance](OIDC_COMPLIANCE_PLAN.md) - OpenID Connect compliance
- [OAuth2 Compliance](OAUTH2_COMPLIANCE_GAP_ANALYSIS.md) - OAuth 2.0 status
- [Scope-Based Claims](SCOPE_BASED_CLAIMS.md) - Claims implementation
- [Auth Time](AUTH_TIME_VERIFICATION.md) - Auth time tracking

### Advanced Features
- [Back-Channel Logout](BACK_CHANNEL_LOGOUT_PLAN.md)
- [Front-Channel Logout](FRONT_CHANNEL_LOGOUT_PLAN.md)
- [RP-Initiated Logout](RP_INITIATED_LOGOUT_PLAN.md)
- [Dynamic Registration](DYNAMIC_REGISTRATION_PLAN.md)
- [Audit Logging](AUDIT_LOGGING_PLAN.md)

### Operations
- [Docker Deployment](DOCKER.md)
- [CI/CD](CI_CD.md)
- [Linting](LINTING_RESOLUTION.md)

## 🔧 GitHub Pages Configuration

This directory is configured for GitHub Pages with:

- **`_config.yml`** - Jekyll configuration
- **`index.md`** - Homepage
- **`TOC.md`** - Complete table of contents
- **`_includes/`** - Reusable navigation components

## 🚀 Enabling GitHub Pages

To enable GitHub Pages for this repository:

1. Go to repository **Settings** → **Pages**
2. Under **Source**, select:
   - Branch: `main`
   - Folder: `/docs`
3. Click **Save**
4. Wait a few minutes for the site to build
5. Visit: `https://prasenjit-net.github.io/openid-golang/`

## 📝 Adding New Documentation

When adding new documentation:

1. Create a new `.md` file in this directory
2. Add Jekyll front matter at the top:
   ```yaml
   ---
   layout: default
   title: Your Page Title
   ---
   ```
3. Add navigation links:
   ```markdown
   [🏠 Home](index.md) | [📚 All Docs](TOC.md) | ...
   
   ---
   ```
4. Add your content
5. Link to your new page from `index.md` or `TOC.md`
6. Commit and push

## 🎨 Theme

The site uses the **Cayman** theme, configured in `_config.yml`.

You can preview locally with:
```bash
# Install Jekyll
gem install bundler jekyll

# Serve locally
cd docs
jekyll serve

# Visit http://localhost:4000/openid-golang/
```

## 📖 Documentation Standards

- Use clear, concise language
- Include code examples where appropriate
- Add navigation at the top of each page
- Use proper markdown formatting
- Include diagrams for complex concepts
- Keep docs up-to-date with code changes

## 🔗 Internal Links

Use relative links for internal documentation:
```markdown
[Link Text](OTHER_DOC.md)
```

GitHub Pages will automatically convert these to proper URLs.

## 📊 Documentation Statistics

- Total docs: 36+ markdown files
- Categories: 6 major sections
- Coverage: Getting started, API, architecture, development, compliance, operations

---

**Maintained by**: OpenID Golang Project Team  
**Last Updated**: November 2025

