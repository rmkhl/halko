// Power setting types matching backend Go structs
export type PowerSettingType = "simple" | "delta";

export interface PowerSettings {
  type?: PowerSettingType;
  power?: number;        // Simple: 0-100
  min_delta?: number;    // Delta control
  max_delta?: number;    // Delta control
}
