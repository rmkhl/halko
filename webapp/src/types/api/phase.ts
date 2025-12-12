// Power setting types matching backend Go structs
export type PowerSettingType = "simple" | "delta" | "pid";

export interface PidSettings {
  kp: number;
  ki: number;
  kd: number;
}

export interface PowerSettings {
  type?: PowerSettingType;
  power?: number;        // Simple: 0-100
  min_delta?: number;    // Delta control
  max_delta?: number;    // Delta control
  pid?: PidSettings;     // PID control
}
