import React from "react";
import { PowerSettings as ApiPowerSettings, PowerSettingType } from "../../types/api";
import { Stack, Typography, TextField, MenuItem, Select, FormControl, InputLabel } from "@mui/material";

const typeLabels: Record<PowerSettingType, string> = {
  simple: "Simple",
  delta: "Delta",
};

const noSpinnerSx = {
  '& input[type=number]': {
    MozAppearance: 'textfield'
  },
  '& input[type=number]::-webkit-outer-spin-button': {
    WebkitAppearance: 'none',
    margin: 0
  },
  '& input[type=number]::-webkit-inner-spin-button': {
    WebkitAppearance: 'none',
    margin: 0
  }
};

interface Props {
  editing?: boolean;
  title: string;
  settings?: ApiPowerSettings;
  onChange: (settings?: ApiPowerSettings) => void;
  // Which control methods this actuator accepts in this step. Steam, for
  // example, may only be modulated against the delta in a heating step. When
  // just one is allowed there is nothing to choose, so the selector is hidden.
  allowedTypes?: PowerSettingType[];
  // Seeds the delta fields when switching to delta control.
  defaultDelta?: { min_delta: number; max_delta: number };
  // What the two delta fields are measured against. An acclimate heater bands
  // the kiln around the step target; everywhere else the band is relative to
  // the material, so the labels have to say which.
  deltaLabels?: { min: string; max: string };
}

export const PowerSettingsComponent: React.FC<Props> = (props) => {
  const {
    editing,
    title,
    settings,
    onChange,
    allowedTypes = ["simple", "delta"],
    defaultDelta = { min_delta: -5, max_delta: 5 },
    deltaLabels = { min: "Min Δ (°C)", max: "Max Δ (°C)" },
  } = props;

  const handleTypeChange = (type: string) => {
    if (type === "simple") {
      onChange({ type: "simple", power: 50 });
    } else if (type === "delta") {
      onChange({ type: "delta", ...defaultDelta });
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


  // Determine control type from settings (matching backend validation logic)
  const getControlType = (): string => {
    if (!settings) return "";

    // Explicit type is set
    if (settings.type) return settings.type;

    // Infer from fields
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
      return `${deltaLabels.min} ${settings?.min_delta}, ${deltaLabels.max} ${settings?.max_delta}`;
    }
    return "Not configured";
  };

  return (
    <Stack direction="row" gap={2} alignItems="center" sx={{ flex: 1 }}>
      <Typography variant="body2" sx={{ width: 100 }}>{title}:</Typography>

      {editing ? (
        <>
          {allowedTypes.length > 1 && (
            <FormControl sx={{ minWidth: 180 }} size="small">
              <InputLabel>Control Type</InputLabel>
              <Select
                value={controlType}
                label="Control Type"
                onChange={(e) => handleTypeChange(e.target.value)}
              >
                {allowedTypes.map((t) => (
                  <MenuItem key={t} value={t}>{typeLabels[t]}</MenuItem>
                ))}
              </Select>
            </FormControl>
          )}

          {controlType === "simple" && (
            <TextField
              label="Power (%)"
              type="number"
              size="small"
              value={settings?.power || 0}
              onChange={(e) => handlePowerChange(Number(e.target.value))}
              inputProps={{ min: 0, max: 100 }}
              sx={{ width: 120, ...noSpinnerSx }}
            />
          )}

          {controlType === "delta" && (
            <>
              <TextField
                label={deltaLabels.min}
                type="number"
                size="small"
                value={settings?.min_delta || 0}
                onChange={(e) => handleDeltaChange("min_delta", Number(e.target.value))}
                sx={{ width: 170, ...noSpinnerSx }}
              />
              <TextField
                label={deltaLabels.max}
                type="number"
                size="small"
                value={settings?.max_delta || 0}
                onChange={(e) => handleDeltaChange("max_delta", Number(e.target.value))}
                sx={{ width: 170, ...noSpinnerSx }}
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
