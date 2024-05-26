import { Program } from "../../types/api";

export const emptyProgram = (): Program => ({
  name: "",
  defaultStepRuntime: 60,
  preheatTo: 200,
  steps: [],
});
