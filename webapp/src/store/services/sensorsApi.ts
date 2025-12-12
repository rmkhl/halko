import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { API_ENDPOINTS } from "../../config/api";

export const sensorApi = createApi({
  reducerPath: "sensorApi",
  baseQuery: fetchBaseQuery({
    baseUrl: API_ENDPOINTS.sensors,
  }),
  endpoints: (builder) => ({
    getTemperatures: builder.query({
      query: () => ({
        url: "/temperatures",
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

export const { useGetTemperaturesQuery } = sensorApi;
