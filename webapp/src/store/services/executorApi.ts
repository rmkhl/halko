import { createApi } from "@reduxjs/toolkit/query/react";
import { fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { fetchSingleQuery, list } from "./queryBuilders";
import { Program } from "../../types/api";

const currentEndpoint = "";
const runningProgramTag = "runningProgram";

export const executorApi = createApi({
  reducerPath: "executorApi",
  baseQuery: fetchBaseQuery({
    baseUrl: "http://localhost:8089/engine/api/v1/running",
  }),
  tagTypes: [runningProgramTag],
  endpoints: (builder) => ({
    getRunningProgram: fetchSingleQuery(
      builder,
      currentEndpoint,
      runningProgramTag
    ),
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
      invalidatesTags: (_, error) =>
        error ? [] : [{ type: runningProgramTag, id: list }],
    }),
  }),
});

export const {
  useGetRunningProgramQuery,
  useStartProgramMutation,
  useStopRunningProgramMutation,
} = executorApi;
