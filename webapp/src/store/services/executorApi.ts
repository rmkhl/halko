import { createApi } from "@reduxjs/toolkit/query/react";
import { fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { fetchSingleQuery } from "./queryBuilders";

const currentEndpoint = "";

export const executorApi = createApi({
  reducerPath: "executorApi",
  baseQuery: fetchBaseQuery({
    baseUrl: "http://localhost:8089/engine/api/v1/running",
  }),
  endpoints: (builder) => ({
    getRunningProgram: fetchSingleQuery(builder, currentEndpoint),
  }),
});

export const { useGetRunningProgramQuery } = executorApi;
