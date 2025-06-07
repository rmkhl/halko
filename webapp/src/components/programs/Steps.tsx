import React from "react";
import { Step as ApiStep } from "../../types/api";
import { Button, Stack, Typography } from "@mui/material";
import { Step } from "./Step";
import { useTranslation } from "react-i18next";
import { v4 as uuidv4 } from "uuid";

interface Props {
  editing?: boolean;
  steps?: ApiStep[];
  onChange: (steps: ApiStep[]) => void;
}

const emptyStep = (): ApiStep => ({
  id: uuidv4(),
  name: "",
  type: "heating",
  targetTemperature: 30,
});

export const Steps: React.FC<Props> = (props) => {
  const { editing, steps, onChange } = props;
  const { t } = useTranslation();

  if (!steps) {
    return null;
  }

  const handleChange = (i: number) => (updatedStep: ApiStep, idx: number) => {
    let updatedSteps = steps.map((step, idx) =>
      idx === i ? updatedStep : step
    );

    if (i !== idx) {
      [updatedSteps[i], updatedSteps[idx]] = [
        updatedSteps[idx],
        updatedSteps[i],
      ];
    }

    onChange(updatedSteps);
  };

  const addStep = (i: number) => () => {
    const cpy = [...steps];
    cpy.splice(i, 0, emptyStep());

    onChange(cpy);
  };

  return (
    <Stack gap={6}>
      <Stack alignItems="center" justifyContent="space-between" direction="row">
        <Typography variant="h4">{t("programs.steps.title")}</Typography>
      </Stack>

      {steps.map((step, i) => (
        <Stack gap={2} key={`step-${step.id}`}>
          <Stack direction="row" justifyContent="space-between">
            <Typography variant="h4" flex={1}>
              {i + 1}
            </Typography>

            <Step
              flex={10}
              key={`step-${i}`}
              editing={editing}
              step={step}
              pos={{
                idx: i,
                isFirst: i === 0,
                isSecond: i === 1,
                isNextToLast: i === steps.length - 2,
                isLast: i === steps.length - 1,
              }}
              onChange={handleChange(i)}
            />
          </Stack>

          {editing && i !== steps.length - 1 && (
            <Stack alignItems="center">
              <Button
                color="success"
                onClick={addStep(i + 1)}
                style={{ width: "2em" }}
              >
                {t("programs.steps.add")}
              </Button>
            </Stack>
          )}
        </Stack>
      ))}
    </Stack>
  );
};
