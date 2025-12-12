import { createApi } from "@reduxjs/toolkit/query/react";
import { fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { fetchSingleQuery, list } from "./queryBuilders";
import { Program } from "../../types/api";
import { API_ENDPOINTS } from "../../config/api";

const currentEndpoint = "";
const runningProgramTag = "runningProgram";
const defaultsTag = "defaults";

export interface EngineDefaults {
  pid_settings: Record<string, { kp: number; ki: number; kd: number }>;
  max_delta_heating: number;
  min_delta_heating: number;
}

export const controlunitApi = createApi({
  reducerPath: "controlunitApi",
  baseQuery: fetchBaseQuery({
    baseUrl: API_ENDPOINTS.controlunit,
  }),
  tagTypes: [runningProgramTag, defaultsTag],
  endpoints: (builder) => ({
    getRunningProgram: fetchSingleQuery(
      builder,
      "/engine/running",
      runningProgramTag
    ),
    getDefaults: builder.query<EngineDefaults, void>({
      query: () => "/engine/defaults",
      providesTags: [defaultsTag],
      transformResponse: (response: { data: EngineDefaults }) => response.data,
    }),
    startProgram: builder.mutation({
      query: (p: Program) => ({
        url: "/engine/running",
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
        url: "/engine/running",
        method: "DELETE",
      }),
      invalidatesTags: (_, error) =>
        error ? [] : [{ type: runningProgramTag, id: list }],
    }),
  }),
});

export const {
  useGetRunningProgramQuery,
  useGetDefaultsQuery,
  useStartProgramMutation,
  useStopRunningProgramMutation,
} = controlunitApi;
