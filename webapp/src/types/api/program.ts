import { PowerSettings } from "./phase";

export type StepType = "heating" | "acclimate" | "cooling";

export interface Step {
  name: string;
  type: StepType;
  temperature_target: number;
  runtime?: string;      // Duration string like "6h", "30m"
  heater?: PowerSettings;
  fan?: PowerSettings;
  humidifier?: PowerSettings;
}

export interface Program {
  name: string;
  steps: Step[];
}
