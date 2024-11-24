import { createApi } from "@reduxjs/toolkit/query/react";
import { fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { fetchQuery, saveMutation } from "./queryBuilders";
import {
  AcclimateStep,
  CoolingStep,
  HeatingStep,
  Program,
  UIProgram,
  defaultAcclimateStep,
  defaultCoolingStep,
  defaultHeatingStep,
} from "../../types/api";

const programsTag = "programs" as const;
const programsEndpoint = "programs";

const programTransformer = (p: Program): UIProgram => ({
  name: p.name,
  heatingStep:
    (p.steps.find((v) => (v.step_type = "heating")) as HeatingStep) ||
    defaultHeatingStep(),
  acclimateStep:
    (p.steps.find((v) => (v.step_type = "acclimate")) as AcclimateStep) ||
    defaultAcclimateStep(),
  coolingStep:
    (p.steps.find((v) => (v.step_type = "cooling")) as CoolingStep) ||
    defaultCoolingStep(),
});

const programSaveTransformer = (p: UIProgram): Program => ({
  name: p.name,
  steps: [p.heatingStep, p.acclimateStep, p.coolingStep],
});

export const configuratorApi = createApi({
  reducerPath: "configuratorApi",
  baseQuery: fetchBaseQuery({
    baseUrl: "http://localhost:8080/api/v1",
  }),
  tagTypes: [programsTag],
  endpoints: (builder) => ({
    getPrograms: fetchQuery(
      builder,
      programsEndpoint,
      programTransformer,
      programsTag
    ),
    saveProgram: saveMutation(
      builder,
      programsEndpoint,
      programSaveTransformer,
      programsTag
    ),
  }),
});

export const { useGetProgramsQuery, useSaveProgramMutation } = configuratorApi;
