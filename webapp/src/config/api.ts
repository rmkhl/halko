/**
 * API Configuration
 *
 * Centralized configuration for all API endpoint base URLs.
 *
 * Environment-based configuration:
 * - Development: Direct localhost access to services
 * - Docker/Production: Relative paths for nginx proxy routing via /api/v1/
 *
 * Set API_PREFIX environment variable to change the base path.
 * Default: empty string (direct access for development)
 * Production: "/api/v1" (for nginx proxy)
 */

const API_PREFIX = process.env.API_PREFIX || '';

interface ApiEndpoints {
  storage: string;
  controlunit: string;
  sensors: string;
}

/**
 * Get API endpoint base URLs
 *
 * In development (no API_PREFIX):
 * - Returns direct localhost URLs for each service
 *
 * In production (API_PREFIX = "/api/v1"):
 * - Returns relative proxy paths that nginx will route to backend services
 */
export const getApiEndpoints = (): ApiEndpoints => {
  if (API_PREFIX) {
    // Production mode: use nginx proxy paths
    // nginx routes /api/v1/{service}/* to the backend service base URL
    return {
      storage: `${API_PREFIX}/storage/storage`,
      controlunit: `${API_PREFIX}/controlunit/engine/running`,
      sensors: `${API_PREFIX}/sensorunit`,
    };
  }

  // Development mode: direct localhost access
  // These match the actual backend service paths from halko.cfg
  return {
    storage: 'http://localhost:8091/storage',
    controlunit: 'http://localhost:8090/engine/running',
    sensors: 'http://localhost:8093',
  };
};

export const API_ENDPOINTS = getApiEndpoints();
