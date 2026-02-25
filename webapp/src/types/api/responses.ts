import { Program } from "./program";

/**
 * Generic API Response wrapper matching Go's APIResponse[T]
 * Used by backend to wrap single-entity responses
 */
export interface APIResponse<T> {
  data: T;
}

/**
 * RTK Query error structure for error handling in mutations/queries
 */
export interface RTKQueryError {
  status?: number;
  data?: {
    error?: string;
  };
  message?: string;
}

/**
 * Temperature readings in Celsius
 * Matches Go's TemperatureStatus struct
 */
export interface TemperatureStatus {
  material: number;
  oven: number;
  delta?: number; // Calculated delta between material and oven
}

/**
 * Power Supply Unit status - power levels in percentage
 * Matches Go's PSUStatus struct
 */
export interface PSUStatus {
  heater: number;
  fan: number;
  humidifier: number;
}

/**
 * Execution status for a running program
 * Matches Go's ExecutionStatus struct
 */
export interface ExecutionStatus {
  program: Program;
  started_at?: number;
  current_step?: string;
  current_step_started_at?: number;
  temperatures?: TemperatureStatus;
  power_status?: PSUStatus;
}

/**
 * Running program response wrapper
 */
export interface RunningProgramResponse {
  data: ExecutionStatus;
}

/**
 * Stored program metadata from storage service
 * Matches Go's StoredProgramInfo struct
 */
export interface StoredProgramInfo {
  name: string;
  last_modified: string;
}

/**
 * Utility type for entities with metadata flags
 * Used in forms and mutations
 */
export type EntityWithMeta<T> = T & { isNew?: boolean };

/**
 * Program with optional fields for form handling
 */
export type ProgramWithOptionalId = Program & { id?: string; isNew?: boolean };
