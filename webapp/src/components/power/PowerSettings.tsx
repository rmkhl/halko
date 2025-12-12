import React from "react";
import { PowerSettings as ApiPowerSettings } from "../../types/api";
import { Stack, Typography, TextField, MenuItem, Select, FormControl, InputLabel } from "@mui/material";

interface Props {
  editing?: boolean;
  title: string;
  settings?: ApiPowerSettings;
  onChange: (settings?: ApiPowerSettings) => void;
}

export const PowerSettingsComponent: React.FC<Props> = (props) => {
  const { editing, title, settings, onChange } = props;

  const handleTypeChange = (type: string) => {
    if (type === "simple") {
      onChange({ type: "simple", power: 50 });
    } else if (type === "delta") {
      onChange({ type: "delta", min_delta: -5, max_delta: 5 });
    } else if (type === "pid") {
      onChange({ type: "pid", pid: { kp: 2.0, ki: 1.0, kd: 0.5 } });
    }
  };

  const handlePowerChange = (power: number) => {
    if (settings) {
      onChange({ ...settings, power });
    }
  };

  const handleDeltaChange = (field: "min_delta" | "max_delta", value: number) => {
    if (settings) {
      onChange({ ...settings, [field]: value });
    }
  };

  const handlePidChange = (field: "kp" | "ki" | "kd", value: number) => {
    if (settings?.pid) {
      onChange({ ...settings, pid: { ...settings.pid, [field]: value } });
    }
  };

  // Determine control type from settings (matching backend validation logic)
  const getControlType = (): string => {
    if (!settings) return "";
    
    // Explicit type is set
    if (settings.type) return settings.type;
    
    // Infer from fields
    if (settings.pid) return "pid";
    if (settings.min_delta !== undefined || settings.max_delta !== undefined) return "delta";
    if (settings.power !== undefined) return "simple";
    
    return "";
  };

  const controlType = getControlType();

  const renderViewMode = () => {
    if (controlType === "simple") {
      return `Power: ${settings?.power}%`;
    }
    if (controlType === "delta") {
      return `Delta: ${settings?.min_delta}°C to ${settings?.max_delta}°C`;
    }
    if (controlType === "pid" && settings?.pid) {
      return `PID: Kp=${settings.pid.kp}, Ki=${settings.pid.ki}, Kd=${settings.pid.kd}`;
    }
    return "Not configured";
  };

  return (
    <Stack direction="row" gap={2} alignItems="center" sx={{ flex: 1 }}>
      <Typography variant="body2" sx={{ width: 100 }}>{title}:</Typography>

      {editing ? (
        <>
          <FormControl sx={{ minWidth: 180 }} size="small">
            <InputLabel>Control Type</InputLabel>
            <Select
              value={controlType}
              label="Control Type"
              onChange={(e) => handleTypeChange(e.target.value)}
            >
              <MenuItem value="simple">Simple</MenuItem>
              <MenuItem value="delta">Delta</MenuItem>
              <MenuItem value="pid">PID</MenuItem>
            </Select>
          </FormControl>

          {controlType === "simple" && (
            <TextField
              label="Power (%)"
              type="number"
              size="small"
              value={settings?.power || 0}
              onChange={(e) => handlePowerChange(Number(e.target.value))}
              inputProps={{ min: 0, max: 100 }}
              sx={{ width: 120 }}
            />
          )}

          {controlType === "delta" && (
            <>
              <TextField
                label="Min Δ (°C)"
                type="number"
                size="small"
                value={settings?.min_delta || 0}
                onChange={(e) => handleDeltaChange("min_delta", Number(e.target.value))}
                sx={{ width: 100 }}
              />
              <TextField
                label="Max Δ (°C)"
                type="number"
                size="small"
                value={settings?.max_delta || 0}
                onChange={(e) => handleDeltaChange("max_delta", Number(e.target.value))}
                sx={{ width: 100 }}
              />
            </>
          )}

          {controlType === "pid" && settings?.pid && (
            <>
              <TextField
                label="Kp"
                type="number"
                size="small"
                value={settings.pid.kp}
                onChange={(e) => handlePidChange("kp", Number(e.target.value))}
                sx={{ width: 80 }}
              />
              <TextField
                label="Ki"
                type="number"
                size="small"
                value={settings.pid.ki}
                onChange={(e) => handlePidChange("ki", Number(e.target.value))}
                sx={{ width: 80 }}
              />
              <TextField
                label="Kd"
                type="number"
                size="small"
                value={settings.pid.kd}
                onChange={(e) => handlePidChange("kd", Number(e.target.value))}
                sx={{ width: 80 }}
              />
            </>
          )}
        </>
      ) : (
        <Typography variant="body2" color="text.secondary">
          {renderViewMode()}
        </Typography>
      )}
    </Stack>
  );
};
