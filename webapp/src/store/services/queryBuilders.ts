import { MutationDefinition, QueryDefinition } from "@reduxjs/toolkit/query";
import {
  ApiBaseQueryFunc,
  ApiBuilder,
  Entity,
  EntityType,
  reducerPath,
} from "./types";

const list = "LIST";

export const fetchQuery = <T>(
  builder: ApiBuilder,
  endpoint: EntityType | string,
  tag?: EntityType
): QueryDefinition<void, ApiBaseQueryFunc, EntityType, T[], reducerPath> =>
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
    providesTags: () => (tag ? [{ type: tag, id: list }] : []),
  });

export const fetchSingleQuery = <T>(
  builder: ApiBuilder,
  endpoint: EntityType | string,
  tag?: EntityType
): QueryDefinition<void, ApiBaseQueryFunc, EntityType, T, reducerPath> =>
  builder.query<T, void>({
    query: () => ({
      url: endpoint,
      responseHandler: (response) => {
        if (!response.ok) {
          return response.text();
        }

        return response.json();
      },
    }),
    providesTags: () => (tag ? [{ type: tag, id: list }] : []),
  });

export const saveMutation = <T extends Entity>(
  builder: ApiBuilder,
  endpoint: EntityType,
  tag: EntityType
): MutationDefinition<T, ApiBaseQueryFunc, EntityType, string, reducerPath> =>
  builder.mutation<string, T>({
    query: (record) => ({
      url: !record.id ? endpoint : endpoint + `/${record.id}`,
      method: !record.id ? "POST" : "PUT",
      body: JSON.stringify(record),
      headers: { "Content-type": "application/json" },
    }),
    invalidatesTags: (_, error) => (error ? [] : [{ type: tag, id: list }]),
  });
