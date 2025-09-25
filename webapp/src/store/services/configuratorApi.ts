import { createApi } from "@reduxjs/toolkit/query/react";
import { fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { fetchQuery, saveMutation } from "./queryBuilders";

const phasesTag = "phases" as const;
const programsTag = "programs" as const;
const phasesEndpoint = "phases";
const programsEndpoint = "programs";

export const configuratorApi = createApi({
  reducerPath: "configuratorApi",
  baseQuery: fetchBaseQuery({
    baseUrl: "http://localhost:8091/storage",
  }),
  tagTypes: [phasesTag, programsTag],
  endpoints: (builder) => ({
    getPhases: fetchQuery(builder, phasesEndpoint, phasesTag),
    savePhase: saveMutation(builder, phasesEndpoint, phasesTag),
    getPrograms: fetchQuery(builder, programsEndpoint, programsTag),
    saveProgram: saveMutation(builder, programsEndpoint, programsTag),
  }),
});

export const {
  useGetPhasesQuery,
  useSavePhaseMutation,
  useGetProgramsQuery,
  useSaveProgramMutation,
} = configuratorApi;
