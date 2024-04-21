import { FetchArgs, createApi } from "@reduxjs/toolkit/query/react";
import { fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { Cycle } from "../../types/api";

const cyclesTag = "cycles" as const;
const list = "LIST";
const cyclesEndpoint = "cycles";

interface Entity {
  id: string;
}

export const configuratorApi = createApi({
  reducerPath: "configuratorApi",
  baseQuery: fetchBaseQuery({
    baseUrl: "http://localhost:8080/api/v1",
  }),
  tagTypes: [cyclesTag],
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
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: cyclesTag, id })),
              { type: cyclesTag, id: list },
            ]
          : [{ type: cyclesTag, id: list }],
    }),
    saveCycle: builder.mutation<string, Cycle>({
      query: (cycle) => ({
        ...entitySaveConfigByEndpoint(cyclesEndpoint, cycle),
      }),
      invalidatesTags: (_, error, cycle) =>
        error ? [] : [{ type: cyclesTag, id: list }],
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

export const { useGetCyclesQuery, useSaveCycleMutation } = configuratorApi;
