import { FetchArgs, createApi } from "@reduxjs/toolkit/query/react";
import { fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { Phase } from "../../types/api";

const phasesTag = "phases" as const;
const list = "LIST";
const phasesEndpoint = "phases";

interface Entity {
  id: string;
}

export const configuratorApi = createApi({
  reducerPath: "configuratorApi",
  baseQuery: fetchBaseQuery({
    baseUrl: "http://localhost:8080/api/v1",
  }),
  tagTypes: [phasesTag],
  endpoints: (builder) => ({
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

export const { useGetPhasesQuery, useSavePhaseMutation } = configuratorApi;
