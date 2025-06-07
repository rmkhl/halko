import { createApi } from "@reduxjs/toolkit/query/react";
import { fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { fetchQuery, saveMutation } from "./queryBuilders";

const programsTag = "programs" as const;
const programsEndpoint = "programs";

export const configuratorApi = createApi({
  reducerPath: "configuratorApi",
  baseQuery: fetchBaseQuery({
    baseUrl: "http://localhost:8080/api/v1",
  }),
  tagTypes: [programsTag],
  endpoints: (builder) => ({
    getPrograms: fetchQuery(builder, programsEndpoint, programsTag),
    saveProgram: saveMutation(builder, programsEndpoint, programsTag),
  }),
});

export const { useGetProgramsQuery, useSaveProgramMutation } = configuratorApi;
