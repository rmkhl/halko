import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { API_ENDPOINTS } from "../../config/api";

// VPN status types
export interface VPNStatus {
  name: string;
  status: string; // "active", "inactive", "failed"
  enabled: boolean;
  tunnel_ip?: string;
}

export interface VPNListResponse {
  data: VPNStatus[];
}

export interface VPNResponse {
  data: VPNStatus;
}

export interface VPNActionResponse {
  data: {
    message: string;
    name: string;
  };
}

export interface PowerActionResponse {
  data: {
    message: string;
  };
}

export const dbusunitApi = createApi({
  reducerPath: "dbusunitApi",
  baseQuery: fetchBaseQuery({
    baseUrl: API_ENDPOINTS.dbusunit,
  }),
  tagTypes: ["vpnList", "vpnStatus"],
  endpoints: (builder) => ({
    // List all VPN connections
    getVPNList: builder.query<VPNListResponse, void>({
      query: () => "/vpn",
      providesTags: ["vpnList"],
    }),

    // Get specific VPN status
    getVPNStatus: builder.query<VPNResponse, string>({
      query: (name) => `/vpn/${name}`,
      providesTags: (_result, _error, name) => [{ type: "vpnStatus", id: name }],
    }),

    // Start VPN connection
    startVPN: builder.mutation<VPNActionResponse, string>({
      query: (name) => ({
        url: `/vpn/${name}/start`,
        method: "POST",
      }),
      invalidatesTags: ["vpnList", "vpnStatus"],
    }),

    // Stop VPN connection
    stopVPN: builder.mutation<VPNActionResponse, string>({
      query: (name) => ({
        url: `/vpn/${name}/stop`,
        method: "POST",
      }),
      invalidatesTags: ["vpnList", "vpnStatus"],
    }),

    // Shutdown system
    shutdownSystem: builder.mutation<PowerActionResponse, void>({
      query: () => ({
        url: "/power/shutdown",
        method: "POST",
      }),
    }),

    // Reboot system
    rebootSystem: builder.mutation<PowerActionResponse, void>({
      query: () => ({
        url: "/power/reboot",
        method: "POST",
      }),
    }),
  }),
});

export const {
  useGetVPNListQuery,
  useGetVPNStatusQuery,
  useStartVPNMutation,
  useStopVPNMutation,
  useShutdownSystemMutation,
  useRebootSystemMutation,
} = dbusunitApi;
