# Admin UI Distribution Files

This directory contains the built files for the admin UI frontend.

## Building the Admin UI

To build the admin UI, run:

```bash
cd frontend
npm install
npm run build
```

The built files will be copied to this directory and embedded into the Go binary.

If this directory is empty (only contains this README and .gitignore), the admin UI 
routes will not be available, but the OpenID server will still function normally.
