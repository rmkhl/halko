# Halko WebApp

React-based user interface for the Halko kiln control system.

## Technology Stack

- **React 18**: UI framework
- **TypeScript**: Type-safe development
- **Material-UI (MUI)**: Component library
- **Redux Toolkit**: State management
- **React Router**: Client-side routing
- **i18next**: Internationalization
- **Parcel**: Build tool and development server

## Development Setup

### Prerequisites

- Node.js 18+ and npm
- The Halko backend services running (controlunit, powerunit, sensorunit)

### Installing Node.js

If you don't have Node.js 18+ installed, the Makefile installs it for you:
`make prepare` (and the webapp targets that depend on it) downloads Node.js
18 standalone binaries into the `.nodejs/` directory in the project root.
No system-wide or user-wide installation is required.

The installed Node.js is automatically used by the webapp Make targets. If you want to use it manually:

```bash
export PATH="$(pwd)/.nodejs/bin:$PATH"
node --version
```

**Note**: The `.nodejs/` directory is git-ignored and project-local. Each developer can have their own installation without conflicts.

### Quick Start

1. Install dependencies:

   ```bash
   cd webapp && npm install
   ```

2. Start the development server (installs dependencies if needed):

   ```bash
   make run-webapp
   # or directly:
   cd webapp && npm start
   ```

   The app will be available at `http://localhost:1234` (default Parcel port) with hot module reloading enabled.

### API Configuration

The webapp uses a centralized API configuration (`src/config/api.ts`) that determines endpoint URLs based on the environment.

**Development Mode** (default):

- Uses direct localhost URLs: `http://localhost:8090`, `http://localhost:8092`, `http://localhost:8093`, `http://localhost:8094`
- Backend services must be running locally
- No environment variables needed

**Production Mode** (nginx proxy):

- Uses relative paths: `/api/v1/controlunit`, `/api/v1/sensorunit`
- Requires nginx to proxy requests to backend services
- Set environment variable: `API_PREFIX=/api/v1`

To configure for production, create a `.env` file:

```bash
cp .env.example .env
# Edit .env and uncomment API_PREFIX=/api/v1
```

Or use the provided `.env.production` file during build:

```bash
npm run build  # Automatically uses .env.production
```

## Production Deployment

### Build for Production

```bash
make build-webapp
```

This creates an optimized production build in `webapp/dist/` with:

- Minified JavaScript and CSS
- Tree-shaking to remove unused code
- Asset optimization
- Hash-based cache busting

The build automatically uses `.env.production` configuration.

### Build Output

The `dist/` directory structure:

```text
dist/
├── index.html          # Entry point
├── *.js               # Bundled JavaScript
├── *.css              # Bundled styles
└── assets/            # Static assets (if any)
```

### Development Mode

The development server provided by Parcel includes:

- Hot module replacement (HMR)
- Source maps for debugging
- Fast incremental builds
- Automatic browser refresh on changes

Start with:

```bash
make run-webapp
```

### Nginx Configuration

The nginx configuration is **automatically generated** using `halkoctl`. The configuration:

- Serves the React SPA
- Handles client-side routing (React Router)
- Proxies API requests to backend services
- Supports WebSocket connections for live updates
- Enables gzip compression
- Sets cache headers for static assets

**Configuration generation:**

For production installation (bare-metal, services on localhost):

```bash
make build-webapp
```

Generates `webapp/nginx-host.conf` using `halko.cfg` which proxies to localhost ports.

**API proxy endpoints:**

- `/api/v1/controlunit/` → `http://localhost:8090` (engine endpoints + WebSocket)
- `/api/v1/storage/` → `http://localhost:8090` (stored programs)
- `/api/v1/powerunit/` → `http://localhost:8092`
- `/api/v1/sensorunit/` → `http://localhost:8093`
- `/api/v1/dbusunit/` → `http://localhost:8094`

## Project Structure

```text
webapp/
├── src/
│   ├── App.tsx                 # Main app component
│   ├── index.tsx              # Entry point
│   ├── components/            # Reusable UI components
│   │   ├── Navigation.tsx
│   │   ├── RouteTabs.tsx
│   │   ├── cycles/           # Cycle management components
│   │   ├── form/             # Form components
│   │   ├── phases/           # Phase configuration
│   │   └── programs/         # Program management
│   ├── hooks/                # Custom React hooks
│   ├── i18n/                 # Internationalization
│   ├── material-ui/          # MUI theme customization
│   ├── store/                # Redux store and slices
│   │   ├── features/         # Redux slices
│   │   └── services/         # RTK Query API services
│   ├── types/                # TypeScript type definitions
│   └── util/                 # Utility functions
├── index.html                # HTML template
├── package.json              # Dependencies and scripts
├── tsconfig.json            # TypeScript configuration
└── nginx-host.conf           # Generated nginx config for production
```

## Make Targets

| Target | Description |
|--------|-------------|
| `prepare` | Check tools; installs Node.js 18 locally to `.nodejs/` if missing |
| `run-webapp` | Start development server with hot reload |
| `build-webapp` | Build for production and generate nginx config (uses `halko.cfg`) |
| `lint-webapp` | Run ESLint and the TypeScript type check |
| `clean-webapp` | Remove build artifacts and dependencies |
| `install-webapp` | Build and install to `/var/www/halko` with nginx config |

## API Integration

The webapp communicates with backend services through RTK Query services in `src/store/services/`:

- `configuratorApi.ts`: Stored program templates
- `controlunitApi.ts`: Program execution control
- `sensorsApi.ts`: Temperature and sensor data
- `powerunitApi.ts`: Power channel status
- `dbusunitApi.ts`: VPN and host power management
- `systemApi.ts`: Aggregated system status
- `queryBuilders.ts`: Common query utilities

API base URLs should be configured based on deployment environment:

- **Development**: Direct service URLs (e.g., `http://localhost:8090`)
- **Production with nginx**: Proxied through `/api/*` endpoints

## Cleaning Up

Remove all build artifacts and dependencies:

```bash
make clean-webapp
```

This removes:

- `dist/` - Production build output
- `node_modules/` - Installed dependencies
- `.parcel-cache/` - Parcel cache files

## Troubleshooting

### Development server won't start

- Ensure Node.js 18+ is installed: `node --version`
- Delete `node_modules` and `.parcel-cache`, then reinstall: `make clean-webapp && cd webapp && npm install`
- Check if port 1234 is already in use

### Build fails

- Check for TypeScript errors: `cd webapp && npm run typecheck`
- Ensure all dependencies are installed: `cd webapp && npm install`
- Check the build output for specific error messages

### API requests fail

- In development, ensure backend services are running
- Check CORS configuration on backend services
- Verify API endpoint URLs in the service configuration
- Check browser console for network errors

## Future Improvements

- Implement automated tests (unit, integration, e2e)
- Add CI/CD pipeline for automated builds and deployments
- Implement service worker for offline capabilities
- Add performance monitoring
