import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { API_ENDPOINTS } from "../../config/api";

// Service status types
export interface ServiceStatus {
  status: "healthy" | "degraded" | "unavailable";
  service: string;
  details: Record<string, unknown>;
}

export interface SystemServicesStatus {
  controlunit: ServiceStatus;
  powerunit: ServiceStatus;
  sensorunit: ServiceStatus;
}

export interface SystemInfo {
  memory_used_mb: number;
  memory_total_mb: number;
  swap_used_mb: number;
  swap_total_mb: number;
  disk_space_mb: number;
}

export interface SystemStatusResponse {
  services: SystemServicesStatus;
  system: SystemInfo;
}

// Hardware status types
export interface ShellyStatus {
  reachable: boolean;
}

export interface HardwareStatusResponse {
  shelly: ShellyStatus;
}

export const systemApi = createApi({
  reducerPath: "systemApi",
  baseQuery: fetchBaseQuery({
    baseUrl: API_ENDPOINTS.controlunit,
  }),
  tagTypes: ["systemStatus", "hardwareStatus"],
  endpoints: (builder) => ({
    getSystemStatus: builder.query<SystemStatusResponse, void>({
      query: () => "/system/status",
      providesTags: ["systemStatus"],
      transformResponse: (response: { data: SystemStatusResponse }) => response.data,
    }),
    getHardwareStatus: builder.query<HardwareStatusResponse, void>({
      query: () => "/system/hardware",
      providesTags: ["hardwareStatus"],
      transformResponse: (response: { data: HardwareStatusResponse }) => response.data,
    }),
  }),
});

export const { useGetSystemStatusQuery, useGetHardwareStatusQuery } = systemApi;
