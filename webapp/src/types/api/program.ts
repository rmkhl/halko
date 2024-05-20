import { Phase } from "./phase";

export interface TemperatureConstraint {
  minimum: number;
  maximum: number;
}

export interface TimeConstraint {
  runtime: number;
}

export interface Step {
  name: string;
  timeConstraint: TimeConstraint;
  temperatureConstraint: TemperatureConstraint;
  heater: Phase;
  fan: Phase;
  humidifier: Phase;
}

export interface Program {
  id: string;
  name: string;
  defaultStepRuntime: number;
  preheatTo: number;
  steps: Step[];
}
