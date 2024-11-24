import { MutationDefinition, QueryDefinition } from "@reduxjs/toolkit/query";
import {
  ApiBaseQueryFunc,
  ApiBuilder,
  Entity,
  EntityType,
  reducerPath,
} from "./types";

export const list = "LIST";

export const fetchQuery = <T, S>(
  builder: ApiBuilder,
  endpoint: EntityType | string,
  dataTransformer: (raw: S) => T,
  tag?: EntityType
): QueryDefinition<void, ApiBaseQueryFunc, EntityType, T[], reducerPath> =>
  builder.query<T[], void>({
    query: () => ({
      url: endpoint,
      responseHandler: async (response) => {
        if (!response.ok) {
          return response.text();
        }

        const raw = (await response.json()) as S[];
        return raw.map((v) => dataTransformer(v));
      },
    }),
    providesTags: () => (tag ? [{ type: tag, id: list }] : []),
  });

export const fetchSingleQuery = <T, S>(
  builder: ApiBuilder,
  endpoint: EntityType | string,
  dataTransformer: (raw: S) => T,
  tag?: EntityType
): QueryDefinition<void, ApiBaseQueryFunc, EntityType, T, reducerPath> =>
  builder.query<T, void>({
    query: () => ({
      url: endpoint,
      responseHandler: async (response) => {
        if (!response.ok) {
          return response.text();
        }

        const raw = (await response.json()) as S;
        return dataTransformer(raw);
      },
    }),
    providesTags: () => (tag ? [{ type: tag, id: list }] : []),
  });

export const saveMutation = <T extends Entity, S>(
  builder: ApiBuilder,
  endpoint: EntityType,
  dataTransformer: (uiData: T) => S,
  tag: EntityType
): MutationDefinition<T, ApiBaseQueryFunc, EntityType, string, reducerPath> =>
  builder.mutation<string, T>({
    query: (record) => ({
      url: endpoint,
      method: "POST",
      body: JSON.stringify(dataTransformer(record)),
      headers: { "Content-type": "application/json" },
    }),
    invalidatesTags: (_, error) => (error ? [] : [{ type: tag, id: list }]),
  });
