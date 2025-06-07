import React from "react";
import { Step as ApiStep, StepType, stepTypes } from "../../types/api";
import { Button, Stack, StackProps } from "@mui/material";
import { useTranslation } from "react-i18next";
import { TextComponent } from "../form/TextComponent";
import ArrowDownwardRoundedIcon from "@mui/icons-material/ArrowDownwardRounded";
import ArrowUpwardRoundedIcon from "@mui/icons-material/ArrowUpwardRounded";
import { SelectionComponent } from "../form/SelectionComponent";
import { NumberComponent } from "../form/NumberComponent";

interface Position {
  idx: number;
  isFirst: boolean;
  isSecond: boolean;
  isNextToLast: boolean;
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
  const { name, type, targetTemperature, heater, humidifier, fan } = step;
  const { t } = useTranslation();

  const handleChange =
    <Key extends keyof ApiStep, Value extends ApiStep[Key]>(key: Key) =>
    (value: Value) =>
      updateStep({ ...step, [key]: value }, pos.idx);

  const handleNudge = (newIdx: number) => updateStep({ ...step }, newIdx);

  const canEdit = !(pos.isFirst || pos.isLast);

  return (
    <Stack gap={3} direction="row" {...rest}>
      <Stack flex={1} gap={2}>
        <TextComponent
          value={name}
          onChange={handleChange("name")}
          editing={editing}
          title={t("programs.steps.name")}
        />

        <SelectionComponent
          editing={editing && canEdit}
          onChange={(s) => handleChange("type")(s as StepType)}
          value={type}
          title="Type"
          options={stepTypes as unknown as string[]}
        />

        <NumberComponent
          value={targetTemperature}
          title="Target temperature"
          onChange={handleChange("targetTemperature")}
          editing={editing}
          min={0}
          max={200}
        >
          °C
        </NumberComponent>

        <NumberComponent
          value={heater?.power}
          title="Heater power"
          onChange={(v) => handleChange("heater")({ power: v, pid: {} })}
          editing={editing}
          min={0}
          max={100}
        >
          %
        </NumberComponent>

        <NumberComponent
          value={humidifier?.power}
          title="Humidifier power"
          onChange={(v) => handleChange("humidifier")({ power: v })}
          editing={editing}
          min={0}
          max={100}
        >
          %
        </NumberComponent>

        <NumberComponent
          value={fan?.power}
          title="Fan power"
          onChange={(v) => handleChange("fan")({ power: v })}
          editing={editing}
          min={0}
          max={100}
        >
          %
        </NumberComponent>
      </Stack>

      {editing && (
        <NudgeColumn pos={pos} onChange={(pos) => handleNudge(pos)} />
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
  const { idx, isFirst, isSecond, isNextToLast, isLast } = pos;

  const handleUpClick = () => {
    onChange(idx - 1);
  };

  const handleDownClick = () => {
    onChange(idx + 1);
  };

  if (isFirst || isLast) {
    return null;
  }

  return (
    <Stack gap={3} justifyContent="center">
      {!isSecond && (
        <Button disabled={idx === 0} onClick={handleUpClick}>
          <ArrowUpwardRoundedIcon />
        </Button>
      )}

      {!isNextToLast && (
        <Button disabled={isLast} onClick={handleDownClick}>
          <ArrowDownwardRoundedIcon />
        </Button>
      )}
    </Stack>
  );
};
