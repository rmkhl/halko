import React from "react";
import { Step as ApiStep } from "../../types/api";
import { Button, Stack, Typography } from "@mui/material";
import { Step } from "./Step";
import { useTranslation } from "react-i18next";

interface Props {
  editing?: boolean;
  steps?: ApiStep[];
  onChange: (steps: ApiStep[]) => void;
}

const emptyStep = (): ApiStep => ({
  name: "",
  type: "heating",
  temperature_target: 100,
  runtime: undefined,
  heater: { type: "simple", power: 100 },
  fan: { type: "simple", power: 100 },
  humidifier: { type: "simple", power: 50 },
});

export const Steps: React.FC<Props> = (props) => {
  const { editing, steps, onChange } = props;
  const { t } = useTranslation();

  if (!steps) {
    return null;
  }

  const handleChange = (i: number) => (updatedStep: ApiStep, idx: number) => {
    const updatedSteps = steps.map((step, idx) =>
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

  const addStep = () => {
    const cpy = [...steps];
    cpy.push(emptyStep());

    onChange(cpy);
  };

  return (
    <Stack gap={6}>
      {editing && (
        <Stack alignItems="center" justifyContent="flex-end" direction="row">
          <Button color="success" onClick={addStep}>
            {t("programs.steps.add")}
          </Button>
        </Stack>
      )}

      {steps.map((step, i) => (
        <Stack key={`step-${i}`} direction="row" justifyContent="space-between">
          <Typography variant="h4" flex={1}>
            {i + 1}
          </Typography>

          <Step
            flex={10}
            editing={editing}
            step={step}
            pos={{ idx: i, isLast: i === steps.length - 1 }}
            onChange={handleChange(i)}
          />
        </Stack>
      ))}
    </Stack>
  );
};
