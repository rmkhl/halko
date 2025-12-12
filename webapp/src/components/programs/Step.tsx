import React from "react";
import { Step as ApiStep, PowerSettings } from "../../types/api";
import { Button, Stack, StackProps, TextField, MenuItem, Select, FormControl, InputLabel, Typography, Box, IconButton, Tooltip } from "@mui/material";
import { TextComponent } from "../form/TextComponent";
import { PowerSettingsComponent } from "../power/PowerSettings";
import ArrowDownwardRoundedIcon from "@mui/icons-material/ArrowDownwardRounded";
import ArrowUpwardRoundedIcon from "@mui/icons-material/ArrowUpwardRounded";
import DeleteIcon from "@mui/icons-material/Delete";
import { useGetDefaultsQuery } from "../../store/services/controlunitApi";

interface Position {
  idx: number;
  isLast: boolean;
}

interface Props extends Omit<StackProps, "onChange"> {
  editing?: boolean;
  step: ApiStep;
  pos: Position;
  onChange: (step: ApiStep, idx: number) => void;
  onDelete?: () => void;
}

export const Step: React.FC<Props> = (props) => {
  const { editing, step, onChange: updateStep, pos, onDelete, ...rest } = props;
  const {
    name,
    type,
    temperature_target,
    runtime,
    heater,
    fan,
    humidifier,
  } = step;
  const { data: defaults } = useGetDefaultsQuery();

  const handleChange =
    <Key extends keyof ApiStep, Value extends ApiStep[Key]>(key: Key) =>
    (value: Value) =>
      updateStep({ ...step, [key]: value }, pos.idx);

  const handleNudge = (newIdx: number) => updateStep({ ...step }, newIdx);

  // Get default heater settings based on step type
  const getDefaultHeater = (): PowerSettings | undefined => {
    if (!defaults) return undefined;

    if (type === "heating") {
      return {
        type: "delta",
        min_delta: defaults.min_delta_heating,
        max_delta: defaults.max_delta_heating,
      };
    } else if (type === "acclimate") {
      const pidSettings = defaults.pid_settings["acclimate"];
      if (pidSettings) {
        return {
          type: "pid",
          pid: pidSettings,
        };
      }
    } else if (type === "cooling") {
      return {
        type: "simple",
        power: 0,
      };
    }
    return undefined;
  };

  // Display heater (use actual or default)
  const displayHeater = heater || getDefaultHeater();

  // Determine current heater power control type
  const getHeaterControlType = (): string => {
    const effectiveHeater = heater || getDefaultHeater();
    const isUsingDefaults = !heater;

    if (!effectiveHeater) {
      return "not configured";
    }

    let controlType = "";

    // Explicit type is set
    if (effectiveHeater.type === "simple") controlType = "constant";
    else if (effectiveHeater.type === "delta") controlType = "delta";
    else if (effectiveHeater.type === "pid") controlType = "PID";
    // Infer from fields (matching backend validation logic)
    else if (effectiveHeater.pid) controlType = "PID";
    else if (effectiveHeater.min_delta !== undefined || effectiveHeater.max_delta !== undefined) controlType = "delta";
    else if (effectiveHeater.power !== undefined) controlType = "constant";
    else return "not configured";

    return isUsingDefaults ? `Default ${controlType}` : controlType;
  };

  return (
    <Stack gap={2} direction="row" {...rest}>
      <Stack flex={1} gap={1.5}>
        <TextComponent
          value={name}
          onChange={handleChange("name")}
          editing={editing}
          title="Step Name"
        />

        <Stack direction="row" gap={2} alignItems="center">
          <Typography variant="body2" sx={{ minWidth: 100 }}>Step Type:</Typography>
          {editing ? (
            <FormControl sx={{ flex: 1 }} size="small">
              <InputLabel>Type</InputLabel>
              <Select
                value={type}
                label="Type"
                onChange={(e) => handleChange("type")(e.target.value as ApiStep["type"])}
              >
                <MenuItem value="heating">Heating</MenuItem>
                <MenuItem value="acclimate">Acclimate</MenuItem>
                <MenuItem value="cooling">Cooling</MenuItem>
              </Select>
            </FormControl>
          ) : (
            <Typography variant="body2" color="text.secondary" sx={{ flex: 1 }}>{type}</Typography>
          )}
        </Stack>

        <Stack direction="row" gap={2} alignItems="center">
          <Box sx={{ minWidth: 100 }} />
          <Stack direction="row" gap={2} alignItems="center" flex={1}>
            <Typography variant="body2" sx={{ width: 100 }}>Target Temp:</Typography>
            {editing ? (
              <TextField
                type="number"
                size="small"
                value={temperature_target}
                onChange={(e) => handleChange("temperature_target")(Number(e.target.value))}
                sx={{
                  width: 80,
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
                }}
                variant="standard"
              />
            ) : (
              <Typography variant="body2" color="text.secondary">{temperature_target}°C</Typography>
            )}
          </Stack>
        </Stack>

        <Stack direction="row" gap={2} alignItems="center">
          <Box sx={{ minWidth: 100 }} />
          <Stack direction="row" gap={2} alignItems="center" flex={1}>
            <Typography variant="body2" sx={{ width: 100 }}>Runtime:</Typography>
            {editing ? (
              <TextField
                size="small"
                value={runtime || ""}
                onChange={(e) => handleChange("runtime")(e.target.value || undefined)}
                placeholder="e.g., 6h"
                sx={{ width: 100 }}
                variant="standard"
              />
            ) : (
              <Typography variant="body2" color="text.secondary">{runtime || "—"}</Typography>
            )}
          </Stack>
        </Stack>

        <Stack direction="row" alignItems="center" sx={{ mt: 0.5, mb: 0.25 }}>
          <Typography variant="body2" sx={{ fontWeight: 500 }}>Power Control:</Typography>
          <Typography variant="caption" color="text.secondary" sx={{ ml: 1 }}>
            {getHeaterControlType()}
          </Typography>
        </Stack>

        <Stack direction="row" gap={2}>
          <Box sx={{ minWidth: 100 }} />
          <Stack flex={1} gap={0.5}>
            <PowerSettingsComponent
              editing={editing}
              title="Heater"
              settings={displayHeater}
              onChange={handleChange("heater")}
            />

            <Stack direction="row" gap={2} alignItems="center">
              <Typography variant="body2" sx={{ width: 100 }}>Fan Power:</Typography>
              {editing ? (
                <TextField
                  type="number"
                  size="small"
                  value={fan?.power ?? 0}
                  onChange={e => handleChange("fan")!({ type: "simple", power: Number(e.target.value) })}
                  inputProps={{ min: 0, max: 100 }}
                  sx={{ width: 100, '& input[type=number]': { MozAppearance: 'textfield' }, '& input[type=number]::-webkit-outer-spin-button': { WebkitAppearance: 'none', margin: 0 }, '& input[type=number]::-webkit-inner-spin-button': { WebkitAppearance: 'none', margin: 0 } }}
                  variant="standard"
                />
              ) : (
                <Typography variant="body2" color="text.secondary">{fan?.power ?? 0}%</Typography>
              )}
            </Stack>

            <Stack direction="row" gap={2} alignItems="center">
              <Typography variant="body2" sx={{ width: 100 }}>Humidifier Power:</Typography>
              {editing ? (
                <TextField
                  type="number"
                  size="small"
                  value={humidifier?.power ?? 0}
                  onChange={e => handleChange("humidifier")!({ type: "simple", power: Number(e.target.value) })}
                  inputProps={{ min: 0, max: 100 }}
                  sx={{ width: 100, '& input[type=number]': { MozAppearance: 'textfield' }, '& input[type=number]::-webkit-outer-spin-button': { WebkitAppearance: 'none', margin: 0 }, '& input[type=number]::-webkit-inner-spin-button': { WebkitAppearance: 'none', margin: 0 } }}
                  variant="standard"
                />
              ) : (
                <Typography variant="body2" color="text.secondary">{humidifier?.power ?? 0}%</Typography>
              )}
            </Stack>
          </Stack>
        </Stack>
      </Stack>

      {editing && (
        <Stack direction="column" gap={1} alignItems="center">
          <NudgeColumn pos={pos} onChange={(pos) => handleNudge(pos)} />
          {onDelete && (
            <Tooltip title="Delete Step">
              <IconButton
                onClick={onDelete}
                color="error"
                size="small"
              >
                <DeleteIcon />
              </IconButton>
            </Tooltip>
          )}
        </Stack>
      )}
    </Stack>
  );
};

interface NudgeColumnProps {
  pos: Position;
  onChange: (pos: number) => void;
}

const NudgeColumn: React.FC<NudgeColumnProps> = (props) => {
  const { pos, onChange } = props;
  const { idx, isLast } = pos;

  const handleUpClick = () => {
    onChange(idx - 1);
  };

  const handleDownClick = () => {
    onChange(idx + 1);
  };

  return (
    <Stack gap={3} justifyContent="center">
      <Button disabled={idx === 0} onClick={handleUpClick} variant="outlined" size="small">
        <ArrowUpwardRoundedIcon />
      </Button>

      <Button disabled={isLast} onClick={handleDownClick} variant="outlined" size="small">
        <ArrowDownwardRoundedIcon />
      </Button>
    </Stack>
  );
};
