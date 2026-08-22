import { PowerSettings } from "./phase";

// The last two are synthesized by the control unit and appear in a running
// program's step list; the editor never offers them.
export type StepType =
  | "heating"
  | "acclimate"
  | "cooling"
  | "equalize"
  | "steam_prewarm";

export interface Step {
  name: string;
  type: StepType;
  temperature_target: number;
  runtime?: string;      // Duration string like "6h", "30m"
  heater?: PowerSettings;
  fan?: PowerSettings;
  steam?: PowerSettings;
}

export interface EqualizeSettings {
  delta?: number;
  steam_prewarm?: boolean;
}

export interface Program {
  name: string;
  description?: string;
  equalize?: EqualizeSettings;
  steps: Step[];
}
