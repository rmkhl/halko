import React from "react";
import {
  AcclimateStep as ApiAcclimateStep,
  defaultAcclimateStep,
} from "../../types/api";
import { Stack, StackProps } from "@mui/material";
import { TextComponent } from "../form/TextComponent";
import { useTranslation } from "react-i18next";
import { TemperatureRangeSlider } from "../form/TemperatureRangeSlider";
import { NumberComponent } from "../form/NumberComponent";

interface Props extends Omit<StackProps, "onChange"> {
  editing?: boolean;
  step?: ApiAcclimateStep;
  onChange: (step: ApiAcclimateStep) => void;
}

export const AcclimateStep: React.FC<Props> = (props) => {
  const {
    editing = false,
    step = defaultAcclimateStep(),
    onChange: updateStep,
    ...rest
  } = props;
  const { name, temperature_target, duration } = step;
  const { t } = useTranslation();

  const handleChange =
    <Key extends keyof ApiAcclimateStep, Value extends ApiAcclimateStep[Key]>(
      key: Key
    ) =>
    (value: Value) =>
      updateStep({ ...step, [key]: value });

  return (
    <Stack gap={3} direction="row" {...rest}>
      <Stack flex={1}>
        <TextComponent
          value={name}
          onChange={handleChange("name")}
          editing={editing}
          title={t("programs.steps.name")}
        />

        <TemperatureRangeSlider
          value={temperature_target}
          editing={editing}
          title="Temperature target"
          onChange={handleChange("temperature_target")}
        />

        <TextComponent
          editing={editing}
          title="Duration (1h15m)"
          value={duration}
          onChange={handleChange("duration")}
        />
      </Stack>
    </Stack>
  );
};
