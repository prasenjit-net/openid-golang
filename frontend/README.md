# Admin UI# React + TypeScript + Vite



React-based admin interface for the OpenID Connect Server.This template provides a minimal setup to get React working in Vite with HMR and some ESLint rules.



## DevelopmentCurrently, two official plugins are available:



### Prerequisites- [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react) uses [Babel](https://babeljs.io/) (or [oxc](https://oxc.rs) when used in [rolldown-vite](https://vite.dev/guide/rolldown)) for Fast Refresh

- Node.js 18+ and npm- [@vitejs/plugin-react-swc](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react-swc) uses [SWC](https://swc.rs/) for Fast Refresh



### Setup## React Compiler

```bash

cd ui/adminThe React Compiler is not enabled on this template because of its impact on dev & build performances. To add it, see [this documentation](https://react.dev/learn/react-compiler/installation).

npm install

```## Expanding the ESLint configuration



### Development ServerIf you are developing a production application, we recommend updating the configuration to enable type-aware lint rules:

```bash

npm run dev```js

```export default defineConfig([

  globalIgnores(['dist']),

This will start the Vite dev server on `http://localhost:3000` with hot module replacement. The dev server is configured to proxy API requests to the backend server at `http://localhost:8080`.  {

    files: ['**/*.{ts,tsx}'],

### Building for Production    extends: [

```bash      // Other configs...

npm run build

```      // Remove tseslint.configs.recommended and replace with this

      tseslint.configs.recommendedTypeChecked,

This builds the UI for production into the `dist/` directory. These files are embedded into the Go binary during the main build process.      // Alternatively, use this for stricter rules

      tseslint.configs.strictTypeChecked,

## Features      // Optionally, add this for stylistic rules

      tseslint.configs.stylisticTypeChecked,

- **Dashboard**: Overview of users, clients, tokens, and login statistics

- **User Management**: Create, list, and delete users      // Other configs...

- **Client Management**: Register OAuth clients, view credentials, manage redirect URIs    ],

- **Settings**: Configure server settings, token TTLs, and key rotation    languageOptions: {

- **Setup Wizard**: Initial configuration for first-time setup      parserOptions: {

- **Authentication**: Secure admin login        project: ['./tsconfig.node.json', './tsconfig.app.json'],

        tsconfigRootDir: import.meta.dirname,

## Technology Stack      },

      // other options...

- **React 18**: UI framework    },

- **TypeScript**: Type-safe development  },

- **Vite**: Build tool with HMR])

- **React Router**: Client-side routing```

- **Axios**: HTTP client for API calls

You can also install [eslint-plugin-react-x](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-x) and [eslint-plugin-react-dom](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-dom) for React-specific lint rules:

## Project Structure

```js

```// eslint.config.js

ui/admin/import reactX from 'eslint-plugin-react-x'

├── public/          # Static assetsimport reactDom from 'eslint-plugin-react-dom'

├── src/

│   ├── components/  # Reusable componentsexport default defineConfig([

│   │   └── Layout.tsx  globalIgnores(['dist']),

│   ├── pages/       # Page components  {

│   │   ├── Dashboard.tsx    files: ['**/*.{ts,tsx}'],

│   │   ├── Users.tsx    extends: [

│   │   ├── Clients.tsx      // Other configs...

│   │   ├── Settings.tsx      // Enable lint rules for React

│   │   ├── Setup.tsx      reactX.configs['recommended-typescript'],

│   │   └── Login.tsx      // Enable lint rules for React DOM

│   ├── App.tsx      # Main app with routing      reactDom.configs.recommended,

│   └── main.tsx     # Entry point    ],

├── index.html    languageOptions: {

├── package.json      parserOptions: {

├── tsconfig.json        project: ['./tsconfig.node.json', './tsconfig.app.json'],

└── vite.config.ts   # Vite configuration        tsconfigRootDir: import.meta.dirname,

```      },

      // other options...

## API Integration    },

  },

The admin UI communicates with the backend through REST API endpoints:])

```

- `GET /api/admin/stats` - Dashboard statistics
- `GET /api/admin/users` - List users
- `POST /api/admin/users` - Create user
- `DELETE /api/admin/users/{id}` - Delete user
- `GET /api/admin/clients` - List OAuth clients
- `POST /api/admin/clients` - Register new client
- `DELETE /api/admin/clients/{id}` - Delete client
- `GET /api/admin/settings` - Get server settings
- `PUT /api/admin/settings` - Update settings
- `GET /api/admin/keys` - List signing keys
- `POST /api/admin/keys/rotate` - Rotate keys
- `GET /api/admin/setup/status` - Check setup status
- `POST /api/admin/setup` - Complete initial setup
- `POST /api/admin/login` - Admin authentication

## Embedding in Go Binary

The built UI files are embedded into the Go binary using Go's `embed` package. See `internal/ui/embed.go` for the implementation.

During production build:
1. React app is built to `ui/admin/dist/`
2. Go `embed` directive includes these files
3. Files are served from the binary at runtime

This creates a single, portable binary with no external dependencies.
