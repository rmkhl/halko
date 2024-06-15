import { Phase } from "./phase";

export interface TemperatureConstraint {
  minimum: number;
  maximum: number;
}

export interface Step {
  name: string;
  timeConstraint: number;
  temperatureConstraint: TemperatureConstraint;
  heater?: Phase;
  fan?: Phase;
  humidifier?: Phase;
}

export interface Program {
  name: string;
  defaultStepRuntime: number;
  preheatTo: number;
  steps: Step[];
}
