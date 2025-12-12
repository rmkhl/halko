import {
  BaseQueryFn,
  EndpointBuilder,
  FetchArgs,
  FetchBaseQueryError,
  FetchBaseQueryMeta,
} from "@reduxjs/toolkit/query";

export type reducerPath = "configuratorApi" | "controlunitApi";

export type EntityType = "programs" | "runningProgram";

export type ApiBuilder = EndpointBuilder<
  BaseQueryFn<
    string | FetchArgs,
    unknown,
    FetchBaseQueryError,
    Record<string, never>,
    FetchBaseQueryMeta
  >,
  EntityType,
  reducerPath
>;

export type ApiBaseQueryFunc = BaseQueryFn<
  string | FetchArgs,
  unknown,
  FetchBaseQueryError,
  Record<string, never>,
  FetchBaseQueryMeta
>;

export interface Entity {
  id: string;
}
