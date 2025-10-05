export interface PowerSetting {
  power: number;
}

export interface PidSettings {
  kp?: number;
  ki?: number;
  kd?: number;
}

export const powerSettingTypes = ["simple", "delta", "pid"] as const;
export type PowerSettingType = (typeof powerSettingTypes)[number];

export interface PowerPidSettings {
  type: PowerSettingType;
  minDelta?: number;
  maxDelta?: number;
  power?: number;
  pid?: PidSettings;
}

export interface TemperatureConstraint {
  minimum: number;
  maximum: number;
}

export const stepTypes = ["heating", "acclimate", "cooling"] as const;
export type StepType = (typeof stepTypes)[number];

export interface Step {
  id: string;
  name: string;
  type: StepType;
  targetTemperature: number;
  runtime: number;
  heater?: PowerPidSettings;
  fan?: PowerSetting;
  humidifier?: PowerSetting;
}

export interface Program {
  name: string;
  steps: Step[];
}
