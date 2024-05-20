import {
  BaseQueryFn,
  FetchArgs,
  FetchBaseQueryError,
  FetchBaseQueryMeta,
  MutationDefinition,
  QueryDefinition,
  createApi,
} from "@reduxjs/toolkit/query/react";
import { fetchBaseQuery, EndpointBuilder } from "@reduxjs/toolkit/query/react";

type entity = "phases" | "programs";
const phasesTag = "phases" as const;
const programsTag = "programs" as const;
const list = "LIST";
const phasesEndpoint = "phases";
const programsEndpoint = "programs";

interface Entity {
  id: string;
}

const fetchQuery = <T>(
  builder: EndpointBuilder<
    BaseQueryFn<
      string | FetchArgs,
      unknown,
      FetchBaseQueryError,
      {},
      FetchBaseQueryMeta
    >,
    entity,
    "configuratorApi"
  >,
  endpoint: entity,
  tag: entity
): QueryDefinition<
  void,
  BaseQueryFn<
    string | FetchArgs,
    unknown,
    FetchBaseQueryError,
    {},
    FetchBaseQueryMeta
  >,
  entity,
  T[],
  "configuratorApi"
> =>
  builder.query<T[], void>({
    query: () => ({
      url: endpoint,
      responseHandler: (response) => {
        if (!response.ok) {
          return response.text();
        }

        return response.json();
      },
    }),
    providesTags: () => [{ type: tag, id: list }],
  });

const saveMutation = <T extends Entity>(
  builder: EndpointBuilder<
    BaseQueryFn<
      string | FetchArgs,
      unknown,
      FetchBaseQueryError,
      {},
      FetchBaseQueryMeta
    >,
    entity,
    "configuratorApi"
  >,
  endpoint: entity,
  tag: entity
): MutationDefinition<
  T,
  BaseQueryFn<
    string | FetchArgs,
    unknown,
    FetchBaseQueryError,
    {},
    FetchBaseQueryMeta
  >,
  entity,
  string,
  "configuratorApi"
> =>
  builder.mutation<string, T>({
    query: (record) => ({
      url: !record.id ? endpoint : endpoint + `/${record.id}`,
      method: !record.id ? "POST" : "PUT",
      body: JSON.stringify(record),
      headers: { "Content-type": "application/json" },
    }),
    invalidatesTags: (_, error) => (error ? [] : [{ type: tag, id: list }]),
  });

export const configuratorApi = createApi({
  reducerPath: "configuratorApi",
  baseQuery: fetchBaseQuery({
    baseUrl: "http://localhost:8080/api/v1",
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
