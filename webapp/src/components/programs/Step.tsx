import React from "react";
import { Step as ApiStep } from "../../types/api";
import { Stack } from "@mui/material";
import { useTranslation } from "react-i18next";
import { TextComponent } from "../form/TextComponent";
import { TimeComponent } from "../form/TimeComponent";
import { TemperatureRangeSlider } from "../form/TemperatureRangeSlider";
import { Phase } from "../phases/Phase";
import { PhaseSelector } from "./PhaseSelector";

interface Props {
  editing?: boolean;
  step: ApiStep;
  onChange: (step: ApiStep) => void;
}

export const Step: React.FC<Props> = (props) => {
  const { editing, step, onChange: updateStep } = props;
  const {
    name,
    timeConstraint,
    temperatureConstraint,
    heater,
    fan,
    humidifier,
  } = step;
  const { t } = useTranslation();

  const handleChange =
    <Key extends keyof ApiStep, Value extends ApiStep[Key]>(key: Key) =>
    (value: Value) =>
      updateStep({ ...step, [key]: value });

  return (
    <Stack gap={3}>
      <TextComponent
        value={name}
        onChange={handleChange("name")}
        editing={editing}
        title={t("programs.steps.name")}
      />

      <TimeComponent
        editing={editing}
        title={t("programs.steps.timeConstraint")}
        value={timeConstraint}
        onChange={handleChange("timeConstraint")}
      />

      <TemperatureRangeSlider
        editing={editing}
        title={t("programs.steps.temperatureConstraint.title")}
        low={temperatureConstraint.minimum}
        high={temperatureConstraint.maximum}
        onChange={(low: number, high: number) => {
          handleChange("temperatureConstraint")({
            minimum: low,
            maximum: high,
          });
        }}
      />

      <PhaseSelector
        editing={editing}
        title={t("programs.steps.heater")}
        phase={heater}
        onChange={handleChange("heater")}
      />

      <PhaseSelector
        editing={editing}
        title={t("programs.steps.fan")}
        phase={fan}
        onChange={handleChange("fan")}
      />

      <PhaseSelector
        editing={editing}
        title={t("programs.steps.humidifier")}
        phase={humidifier}
        onChange={handleChange("humidifier")}
      />
    </Stack>
  );
};
