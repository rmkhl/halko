import { createApi } from "@reduxjs/toolkit/query/react";
import { fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { fetchQuery, saveMutation } from "./queryBuilders";
import { API_ENDPOINTS } from "../../config/api";

const programsTag = "programs" as const;
const programsEndpoint = "/programs";

export const configuratorApi = createApi({
  reducerPath: "configuratorApi",
  baseQuery: fetchBaseQuery({
    baseUrl: API_ENDPOINTS.storage,
  }),
  tagTypes: [programsTag],
  endpoints: (builder) => ({
    getPrograms: fetchQuery(builder, programsEndpoint, programsTag),
    getProgram: builder.query({
      query: (name: string) => `${programsEndpoint}/${encodeURIComponent(name)}`,
      providesTags: [programsTag],
    }),
    saveProgram: saveMutation(builder, programsEndpoint, programsTag),
    deleteProgram: builder.mutation<void, string>({
      query: (name: string) => ({
        url: `${programsEndpoint}/${encodeURIComponent(name)}`,
        method: "DELETE",
      }),
      invalidatesTags: [programsTag],
    }),
  }),
});

export const {
  useGetProgramsQuery,
  useGetProgramQuery,
  useSaveProgramMutation,
  useDeleteProgramMutation,
} = configuratorApi;
