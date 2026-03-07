import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { API_ENDPOINTS } from "../../config/api";
import { APIResponse, PowerStatusResponse } from "../../types/api/responses";

export const powerunitApi = createApi({
  reducerPath: "powerunitApi",
  baseQuery: fetchBaseQuery({
    baseUrl: API_ENDPOINTS.powerunit,
  }),
  endpoints: (builder) => ({
    getPowerStatus: builder.query<APIResponse<PowerStatusResponse>, void>({
      query: () => ({
        url: "/power",
        responseHandler: (response) => {
          if (!response.ok) {
            return response.text();
          }

          return response.json();
        },
      }),
    }),
  }),
});

export const { useGetPowerStatusQuery } = powerunitApi;
