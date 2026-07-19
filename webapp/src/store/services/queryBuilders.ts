import { EndpointBuilder } from "@reduxjs/toolkit/query";
import { ApiBaseQueryFunc } from "./types";
import { EntityWithMeta } from "../../types/api";

export const list = "LIST";

export const fetchQuery = <T, Tag extends string, Path extends string>(
  builder: EndpointBuilder<ApiBaseQueryFunc, Tag, Path>,
  endpoint: string,
  tag?: Tag
) =>
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

export const fetchSingleQuery = <T, Tag extends string, Path extends string>(
  builder: EndpointBuilder<ApiBaseQueryFunc, Tag, Path>,
  endpoint: string,
  tag?: Tag
) =>
  builder.query<T, void>({
    query: () => ({
      url: endpoint,
      responseHandler: (response) => {
        // Handle 204 No Content - return null
        if (response.status === 204) {
          return Promise.resolve(null);
        }

        if (!response.ok) {
          return response.text();
        }

        return response.json();
      },
    }),
    providesTags: () => (tag ? [{ type: tag, id: list }] : []),
  });

export const saveMutation = <T extends { name: string }, Tag extends string, Path extends string>(
  builder: EndpointBuilder<ApiBaseQueryFunc, Tag, Path>,
  endpoint: string,
  tag: Tag
) =>
  builder.mutation<string, EntityWithMeta<T>>({
    query: (record) => {
      // Use explicit isNew flag if present
      const isNew = record.isNew === true;
      // Remove isNew from payload (eslint: _isNew is intentionally unused)
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { isNew: _isNew, ...payload } = record;
      return {
        url: isNew ? endpoint : `${endpoint}/${encodeURIComponent(record.name)}`,
        method: "POST",
        body: JSON.stringify(payload),
        headers: { "Content-type": "application/json" },
      };
    },
    invalidatesTags: (_, error) => (error ? [] : [{ type: tag, id: list }]),
  });
