import { createApi } from "@reduxjs/toolkit/query/react";
import { fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { fetchSingleQuery } from "./queryBuilders";
import { Program } from "../../types/api";

const currentEndpoint = "";

export const executorApi = createApi({
  reducerPath: "executorApi",
  baseQuery: fetchBaseQuery({
    baseUrl: "http://localhost:8089/engine/api/v1/running",
  }),
  endpoints: (builder) => ({
    getRunningProgram: fetchSingleQuery(builder, currentEndpoint),
    startProgram: builder.mutation({
      query: (p: Program) => ({
        url: "",
        method: "POST",
        body: JSON.stringify({
          name: p.name,
          default_step_time: p.defaultStepRuntime,
          steps: p.steps.map((s) => ({
            name: s.name,
            time_constraint: s.timeConstraint,
            temperature_constraint: s.temperatureConstraint,
            heater: s.heater,
            fan: s.fan,
            humidifier: s.humidifier,
          })),
        }),
        headers: { "Content-type": "application/json" },
      }),
    }),
    stopRunningProgram: builder.mutation({
      query: () => ({
        url: "",
        method: "DELETE",
      }),
    }),
  }),
});

export const {
  useGetRunningProgramQuery,
  useStartProgramMutation,
  useStopRunningProgramMutation,
} = executorApi;
