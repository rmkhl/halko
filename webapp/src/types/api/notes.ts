import { TemperatureStatus } from "./responses";

// An operator's observation, recorded while a run was in progress. Everything
// but the text is stamped by the control unit at the moment it was taken.
export interface RunNote {
  time: number; // unix seconds
  step: string;
  temperatures: TemperatureStatus;
  text: string;
}
