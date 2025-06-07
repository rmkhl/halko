import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";

export const sensorApi = createApi({
  reducerPath: "sensorApi",
  baseQuery: fetchBaseQuery({
    baseUrl: "http://localhost:8088/sensors",
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
