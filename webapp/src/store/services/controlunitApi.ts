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

export interface RunHistory {
  name: string;
  state: "completed" | "failed" | "canceled" | "running" | "pending" | "unknown";
  started_at?: number;
  completed_at?: number;
}

export interface ExecutedProgram {
  name: string;
  state: string;
  started_at?: number;
  completed_at?: number;
  program: Program;
}

export const controlunitApi = createApi({
  reducerPath: "controlunitApi",
  baseQuery: fetchBaseQuery({
    baseUrl: API_ENDPOINTS.controlunit,
  }),
  tagTypes: [runningProgramTag, defaultsTag, "history"],
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
          steps: p.steps.map((s) => {
            // Only include fields present in the step, do not use defaults
            const step: any = {
              name: s.name,
              type: s.type,
              temperature_target: s.temperature_target,
            };
            if (s.runtime !== undefined) step.runtime = s.runtime;
            if (s.heater !== undefined) step.heater = s.heater;
            if (s.fan !== undefined) step.fan = s.fan;
            if (s.humidifier !== undefined) step.humidifier = s.humidifier;
            return step;
          }),
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
    getExecutionHistory: builder.query<RunHistory[], void>({
      query: () => "/engine/history",
      providesTags: ["history"],
      transformResponse: (response: { data: RunHistory[] }) => response.data,
    }),
    getExecutionLog: builder.query<string, string>({
      query: (name) => ({
        url: `/engine/history/${encodeURIComponent(name)}/log`,
        responseHandler: (response) => response.text(),
      }),
    }),
    deleteExecution: builder.mutation<void, string>({
      query: (name) => ({
        url: `/engine/history/${encodeURIComponent(name)}`,
        method: "DELETE",
      }),
      invalidatesTags: ["history"],
    }),
  }),
});

export const {
  useGetRunningProgramQuery,
  useGetDefaultsQuery,
  useStartProgramMutation,
  useStopRunningProgramMutation,
  useGetExecutionHistoryQuery,
  useGetExecutionLogQuery,
  useDeleteExecutionMutation,
} = controlunitApi;
