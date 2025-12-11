import {
  BaseQueryFn,
  EndpointBuilder,
  FetchArgs,
  FetchBaseQueryError,
  FetchBaseQueryMeta,
} from "@reduxjs/toolkit/query";

export type reducerPath = "configuratorApi" | "controlunitApi";

export type EntityType = "phases" | "programs" | "runningProgram";

export type ApiBuilder = EndpointBuilder<
  BaseQueryFn<
    string | FetchArgs,
    unknown,
    FetchBaseQueryError,
    {},
    FetchBaseQueryMeta
  >,
  EntityType,
  reducerPath
>;

export type ApiBaseQueryFunc = BaseQueryFn<
  string | FetchArgs,
  unknown,
  FetchBaseQueryError,
  {},
  FetchBaseQueryMeta
>;

export interface Entity {
  id: string;
}
