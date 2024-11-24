import React from "react";
import {
  HeatingStep as ApiHeatingStep,
  defaultHeatingStep,
} from "../../types/api";
import { Stack, StackProps } from "@mui/material";
import { TextComponent } from "../form/TextComponent";
import { useTranslation } from "react-i18next";
import { TemperatureRangeSlider } from "../form/TemperatureRangeSlider";

interface Props extends Omit<StackProps, "onChange"> {
  editing?: boolean;
  step?: ApiHeatingStep;
  onChange: (step: ApiHeatingStep) => void;
}

export const HeatingStep: React.FC<Props> = (props) => {
  const {
    editing = false,
    step = defaultHeatingStep(),
    onChange: updateStep,
    ...rest
  } = props;
  const { name, temperature_target } = step;
  const { t } = useTranslation();

  const handleChange =
    <Key extends keyof ApiHeatingStep, Value extends ApiHeatingStep[Key]>(
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
      </Stack>
    </Stack>
  );
};
