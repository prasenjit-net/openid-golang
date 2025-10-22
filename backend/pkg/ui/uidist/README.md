# UI Distribution Directory

This directory is used to store the compiled frontend assets for the Admin UI.

## Purpose

The `embed.go` file in the parent directory embeds this `uidist` directory into the Go binary using `//go:embed uidist`. This allows the server to serve the admin UI without requiring external files.

## Contents

When the frontend is built, the compiled assets will be placed here:
- `index.html` - Main HTML file
- `assets/` - JavaScript, CSS, and other static assets

## Build Frontend

To build the frontend and populate this directory:

```bash
cd frontend
npm install
npm run build
```

The build process (configured in `vite.config.ts`) outputs to `../backend/pkg/ui/uidist`.

## Note

This directory is tracked in git (not ignored) because the embed directive requires it to exist. However, the actual build artifacts inside are typically ignored by `.gitignore` patterns.
