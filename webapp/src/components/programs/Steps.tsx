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
  timeConstraint: 360,
  temperatureConstraint: {
    minimum: 50,
    maximum: 100,
  },
});

export const Steps: React.FC<Props> = (props) => {
  const { editing, steps, onChange } = props;
  const { t } = useTranslation();

  if (!steps) {
    return null;
  }

  const handleChange = (i: number) => (updatedStep: ApiStep) => {
    onChange(steps.map((step, idx) => (idx === i ? updatedStep : step)));
  };

  const addStep = () => {
    const cpy = [...steps];
    cpy.push(emptyStep());

    onChange(cpy);
  };

  return (
    <Stack gap={6}>
      <Stack alignItems="center" justifyContent="space-between" direction="row">
        <Typography variant="h6">{t("programs.steps.title")}</Typography>

        {editing && (
          <Button color="success" onClick={addStep}>
            {t("programs.steps.add")}
          </Button>
        )}
      </Stack>

      {steps.map((step, i) => (
        <Step
          key={`step-${i}`}
          editing={editing}
          step={step}
          onChange={handleChange(i)}
        />
      ))}
    </Stack>
  );
};
