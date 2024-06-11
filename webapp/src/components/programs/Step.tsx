import React from "react";
import { Step as ApiStep } from "../../types/api";
import { Button, Stack, StackProps } from "@mui/material";
import { useTranslation } from "react-i18next";
import { TextComponent } from "../form/TextComponent";
import { TimeComponent } from "../form/TimeComponent";
import { TemperatureRangeSlider } from "../form/TemperatureRangeSlider";
import { PhaseSelector } from "./PhaseSelector";
import ArrowDownwardRoundedIcon from "@mui/icons-material/ArrowDownwardRounded";
import ArrowUpwardRoundedIcon from "@mui/icons-material/ArrowUpwardRounded";

interface Position {
  idx: number;
  isLast: boolean;
}

interface Props extends Omit<StackProps, "onChange"> {
  editing?: boolean;
  step: ApiStep;
  pos: Position;
  onChange: (step: ApiStep, idx: number) => void;
}

export const Step: React.FC<Props> = (props) => {
  const { editing, step, onChange: updateStep, pos, ...rest } = props;
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
      updateStep({ ...step, [key]: value }, pos.idx);

  const handleNudge = (newIdx: number) => updateStep({ ...step }, newIdx);

  return (
    <Stack gap={3} direction="row" {...rest}>
      <Stack flex={1}>
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

      <NudgeColumn pos={pos} onChange={(pos) => handleNudge(pos)} />
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
      <Button disabled={idx === 0} onClick={handleUpClick}>
        <ArrowUpwardRoundedIcon />
      </Button>

      <Button disabled={isLast} onClick={handleDownClick}>
        <ArrowDownwardRoundedIcon />
      </Button>
    </Stack>
  );
};
