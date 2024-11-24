import { AcclimateStep } from "../../components/programs/AcclimateStep";

export interface PowerSetting {
  power: number;
}

export interface PIDSettings {
  kp: number;
  ki: number;
  kd: number;
}

export interface PowerPIDSettings extends Partial<PowerSetting> {
  max_delta?: number;
  pid?: PIDSettings;
}

export interface HeatingPIDSettings
  extends Required<Pick<PowerPIDSettings, "max_delta" | "pid">> {}

export interface CoolingPowerPIDSettings
  extends Required<Pick<PowerPIDSettings, "power">> {}

export interface HeatingStep
  extends Required<Pick<Step, "name" | "temperature_target">> {
  step_type: "heating";
  heater: HeatingPIDSettings;
  fan: Required<PowerSetting>;
  humidifier: Required<PowerSetting>;
}

export interface AcclimateStep
  extends Required<Pick<Step, "name" | "duration" | "temperature_target">> {
  step_type: "acclimate";
  heater: HeatingPIDSettings;
  fan: Required<PowerSetting>;
  humidifier: Required<PowerSetting>;
}

export interface CoolingStep
  extends Required<Pick<Step, "name" | "temperature_target">> {
  step_type: "cooling";
  heater: CoolingPowerPIDSettings;
  fan: Required<PowerSetting>;
  humidifier: Required<PowerSetting>;
}

export interface Step {
  name: string;
  step_type: "heating" | "cooling" | "acclimate";
  duration?: string;
  temperature_target?: number;
  heater?: PowerPIDSettings;
  fan?: PowerSetting;
  humidifier?: PowerSetting;
}

export interface Program {
  name: string;
  steps: Step[];
}

export interface UIProgram {
  name: string;
  heatingStep: HeatingStep;
  acclimateStep: AcclimateStep;
  coolingStep: CoolingStep;
}

const defaultPID = (): PIDSettings => ({
  kp: 1,
  ki: 1,
  kd: 1,
});

const defaultHeat = (): HeatingPIDSettings => ({
  max_delta: 5,
  pid: defaultPID(),
});

export const defaultHeatingStep = (): HeatingStep => ({
  name: "heat",
  step_type: "heating",
  temperature_target: 200,
  heater: defaultHeat(),
  fan: { power: 50 },
  humidifier: { power: 50 },
});

export const defaultAcclimateStep = (): AcclimateStep => ({
  name: "acclimate",
  step_type: "acclimate",
  temperature_target: 200,
  duration: "10h",
  heater: defaultHeat(),
  fan: { power: 50 },
  humidifier: { power: 50 },
});

export const defaultCoolingStep = (): CoolingStep => ({
  name: "cooling",
  step_type: "cooling",
  temperature_target: 20,
  heater: { power: 0 },
  fan: { power: 100 },
  humidifier: { power: 0 },
});

export const defaultProgram = (): UIProgram => ({
  name: "New program",
  heatingStep: defaultHeatingStep(),
  acclimateStep: defaultAcclimateStep(),
  coolingStep: defaultCoolingStep(),
});
