import React from "react";
import {
  CoolingStep as ApiCoolingStep,
  defaultCoolingStep,
} from "../../types/api";
import { Stack, StackProps } from "@mui/material";
import { TextComponent } from "../form/TextComponent";
import { TemperatureRangeSlider } from "../form/TemperatureRangeSlider";
import { useTranslation } from "react-i18next";

interface Props extends Omit<StackProps, "onChange"> {
  editing?: boolean;
  step?: ApiCoolingStep;
  onChange: (step: ApiCoolingStep) => void;
}

export const CoolingStep: React.FC<Props> = (props) => {
  const {
    editing = false,
    step = defaultCoolingStep(),
    onChange: updateStep,
    ...rest
  } = props;
  const { name, temperature_target } = step;
  const { t } = useTranslation();

  const handleChange =
    <Key extends keyof ApiCoolingStep, Value extends ApiCoolingStep[Key]>(
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
          range={[20, 100]}
        />
      </Stack>
    </Stack>
  );
};
