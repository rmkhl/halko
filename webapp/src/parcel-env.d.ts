/**
 * Environment variable type definitions for Parcel bundler
 *
 * This file provides TypeScript type safety for environment variables
 * accessed via process.env in the application.
 *
 * Parcel automatically injects environment variables at build time:
 * - Variables prefixed with API_ are available via process.env.API_*
 * - Values are replaced at build time (not runtime)
 * - Can be set via .env files or shell environment
 *
 * Usage in code:
 *   const prefix = process.env.API_PREFIX || '';
 *
 * Setting values:
 *   - Development: Create .env.development file
 *   - Production: Create .env.production file or set shell variables
 */

declare namespace NodeJS {
  interface ProcessEnv {
    /**
     * API route prefix for proxied requests
     * - Development: '' (empty, direct localhost access)
     * - Production: '/api/v1' (nginx proxy routing)
     */
    readonly API_PREFIX?: string;
  }
}
