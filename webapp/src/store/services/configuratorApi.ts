import { FetchArgs, createApi } from "@reduxjs/toolkit/query/react";
import { fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { Cycle, Phase } from "../../types/api";

const cyclesTag = "cycles" as const;
const phasesTag = "phases" as const;
const list = "LIST";
const cyclesEndpoint = "cycles";
const phasesEndpoint = "phases";

interface Entity {
  id: string;
}

export const configuratorApi = createApi({
  reducerPath: "configuratorApi",
  baseQuery: fetchBaseQuery({
    baseUrl: "http://localhost:8080/api/v1",
  }),
  tagTypes: [cyclesTag, phasesTag],
  endpoints: (builder) => ({
    getCycles: builder.query<Cycle[], void>({
      query: () => ({
        url: cyclesEndpoint,
        responseHandler: (response) => {
          if (!response.ok) {
            return response.text();
          }

          return response.json();
        },
      }),
      providesTags: () => [{ type: cyclesTag, id: list }],
    }),
    saveCycle: builder.mutation<string, Cycle>({
      query: (cycle) => ({
        ...entitySaveConfigByEndpoint(cyclesEndpoint, cycle),
      }),
      invalidatesTags: (_, error) =>
        error ? [] : [{ type: cyclesTag, id: list }],
    }),
    getPhases: builder.query<Phase[], void>({
      query: () => ({
        url: phasesEndpoint,
        responseHandler: (response) => {
          if (!response.ok) {
            return response.text();
          }

          return response.json();
        },
      }),
      providesTags: () => [{ type: phasesTag, id: list }],
    }),
    savePhase: builder.mutation<string, Phase>({
      query: (phase) => ({
        ...entitySaveConfigByEndpoint(phasesEndpoint, phase),
      }),
      invalidatesTags: (_, error) =>
        error ? [] : [{ type: phasesTag, id: list }],
    }),
  }),
});

const entitySaveConfigByEndpoint = (
  endpoint: string,
  entity: Entity
): FetchArgs => ({
  url: !entity.id ? endpoint : endpoint + `/${entity.id}`,
  method: !entity.id ? "POST" : "PUT",
  body: JSON.stringify(entity),
  headers: { "Content-type": "application/json" },
});

export const {
  useGetCyclesQuery,
  useSaveCycleMutation,
  useGetPhasesQuery,
  useSavePhaseMutation,
} = configuratorApi;
