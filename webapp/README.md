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
- The Halko backend services running (executor, storage, powerunit, sensorunit)

### Installing Node.js

If you don't have Node.js 18+ installed, you can install it locally in the project using the provided Make target:

```bash
make webapp-install-node
```

This will:
1. Download Node.js 18.20.5 standalone binaries
2. Extract them to `.nodejs/` directory in the project root
3. No system-wide or user-wide installation required

The installed Node.js is automatically used by all `webapp-*` Make targets. If you want to use it manually:

```bash
export PATH="$(pwd)/.nodejs/bin:$PATH"
node --version
```

**Note**: The `.nodejs/` directory is git-ignored and project-local. Each developer can have their own installation without conflicts.

### Quick Start

1. Install dependencies:
   ```bash
   make webapp-install
   # or directly:
   cd webapp && npm install
   ```

2. Start the development server:
   ```bash
   make webapp-dev
   # or directly:
   cd webapp && npm start
   ```

   The app will be available at http://localhost:1234 (default Parcel port) with hot module reloading enabled.

### API Configuration

The webapp uses a centralized API configuration (`src/config/api.ts`) that determines endpoint URLs based on the environment.

**Development Mode** (default):
- Uses direct localhost URLs: `http://localhost:8090`, `http://localhost:8091`, `http://localhost:8088`
- Backend services must be running locally
- No environment variables needed

**Production Mode** (Docker/nginx proxy):
- Uses relative paths: `/api/v1/executor`, `/api/v1/storage`, `/api/v1/sensorunit`
- Requires nginx to proxy requests to backend services
- Set environment variable: `VITE_API_PREFIX=/api/v1`

To configure for production, create a `.env` file:
```bash
cp .env.example .env
# Edit .env and uncomment VITE_API_PREFIX=/api/v1
```

Or use the provided `.env.production` file during build:
```bash
npm run build  # Automatically uses .env.production
```

### Development Mode

The development server provided by Parcel includes:
- Hot module replacement (HMR)
- Source maps for debugging
- Fast incremental builds
- Automatic browser refresh on changes

## Production Build

### Build for Production

```bash
make webapp-build
```

This creates an optimized production build in `webapp/dist/` with:
- Minified JavaScript and CSS
- Tree-shaking to remove unused code
- Asset optimization
- Hash-based cache busting

### Build Output

The `dist/` directory structure:
```
dist/
├── index.html          # Entry point
├── *.js               # Bundled JavaScript
├── *.css              # Bundled styles
└── assets/            # Static assets (if any)
```

## Docker Deployment

### Build Docker Image

```bash
make webapp-docker-build
```

This creates a multi-stage Docker image that:
1. Generates nginx.conf using `halkoctl` with `halko-docker.cfg`
2. Builds the React app using Node.js
3. Serves the static files using nginx Alpine

### Docker Compose

The webapp is integrated into the main `docker-compose.yml`. Building all images:

```bash
# Build all services including webapp (recommended)
make images

# Or start with docker-compose (will build if needed)
docker-compose up -d
```

Access the webapp at http://localhost:8080

The `make images` target automatically:
- Builds all Go binaries
- Builds the webapp production bundle
- Generates `nginx.conf` using `halkoctl` and `halko-docker.cfg`
- Builds all Docker images including the webapp

### Nginx Configuration

The nginx configuration is **automatically generated** during Docker builds using `halkoctl`. The configuration:
- Serves the React SPA
- Handles client-side routing (React Router)
- Proxies API requests to backend services (using Docker service names)
- Enables gzip compression
- Sets cache headers for static assets

**Manual configuration generation:**

For standard installation (services on localhost):
```bash
make webapp-nginx-config
```
Generates `webapp/nginx-standard.conf` using `halko.cfg` which proxies to localhost ports.

For Docker deployment (services by name):
```bash
make webapp-nginx-docker-config
```
Generates `webapp/nginx.conf` using `halko-docker.cfg` which proxies to Docker service names.

API proxy endpoints (Docker mode):
- `/api/v1/executor/` → `http://executor:8089`
- `/api/v1/storage/` → `http://storage:8091`
- `/api/v1/powerunit/` → `http://powerunit:8090`
- `/api/v1/sensorunit/` → `http://simulator:8093`

## Project Structure

```
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
├── Dockerfile               # Multi-stage Docker build
└── nginx.conf               # Nginx server configuration
```

## Make Targets

| Target | Description |
|--------|-------------|
| `webapp-install-node` | Install Node.js 18 locally in `.nodejs/` (project-specific) |
| `webapp-check-node` | Verify Node.js 18+ and npm are installed |
| `webapp-install` | Install npm dependencies |
| `webapp-dev` | Start development server with hot reload |
| `webapp-build` | Build for production |
| `webapp-clean` | Remove build artifacts and dependencies |
| `webapp-docker-build` | Build Docker image |
| `webapp-nginx-config` | Generate nginx config for standard installation (uses `halko.cfg`) |
| `webapp-nginx-docker-config` | Generate nginx config for Docker deployment (uses `halko-docker.cfg`) |

## API Integration

The webapp communicates with backend services through RTK Query services in `src/store/services/`:

- `configuratorApi.ts`: Not yet implemented (placeholder)
- `executorApi.ts`: Program execution control
- `sensorsApi.ts`: Temperature and sensor data
- `queryBuilders.ts`: Common query utilities

API base URLs should be configured based on deployment environment:
- **Development**: Direct service URLs (e.g., `http://localhost:8090`)
- **Docker**: Service names as hostnames (e.g., `http://executor:8090`)
- **Production with nginx**: Proxied through `/api/*` endpoints

## Cleaning Up

Remove all build artifacts and dependencies:

```bash
make webapp-clean
```

This removes:
- `dist/` - Production build output
- `node_modules/` - Installed dependencies
- `.parcel-cache/` - Parcel cache files

## Troubleshooting

### Development server won't start
- Ensure Node.js 18+ is installed: `node --version`
- Delete `node_modules` and `.parcel-cache`, then reinstall: `make webapp-clean && make webapp-install`
- Check if port 1234 is already in use

### Build fails
- Check for TypeScript errors: `cd webapp && npx tsc --noEmit`
- Ensure all dependencies are installed: `make webapp-install`
- Check the build output for specific error messages

### Docker image issues
- Ensure the production build works locally first: `make webapp-build`
- Check Docker logs: `docker-compose logs webapp`
- Verify nginx configuration syntax: `docker exec halko_webapp nginx -t`

### API requests fail
- In development, ensure backend services are running
- Check CORS configuration on backend services
- Verify API endpoint URLs in the service configuration
- Check browser console for network errors

## Future Improvements

- Add environment variable configuration for API endpoints
- Implement automated tests (unit, integration, e2e)
- Add TypeScript strict mode compliance
- Set up ESLint and Prettier for code quality
- Add CI/CD pipeline for automated builds and deployments
- Implement service worker for offline capabilities
- Add performance monitoring
