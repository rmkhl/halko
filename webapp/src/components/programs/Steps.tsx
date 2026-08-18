import React from "react";
import { Step as ApiStep } from "../../types/api";
import { Button, Stack, Typography, Paper, Box } from "@mui/material";
import { Step } from "./Step";
import AddIcon from "@mui/icons-material/Add";

interface Props {
  editing?: boolean;
  steps?: ApiStep[];
  onChange: (steps: ApiStep[]) => void;
}

// Left without a heater, fan or steam on purpose: the control unit fills each
// one in from its configured defaults, and Step renders those until they are
// overridden. Naming them here would duplicate the configuration and go wrong
// as soon as the step type is changed.
const emptyStep = (): ApiStep => ({
  name: "",
  type: "heating",
  temperature_target: 0,
  runtime: undefined,
  heater: undefined,
  fan: undefined,
  steam: undefined,
});

export const Steps: React.FC<Props> = (props) => {
  const { editing, steps, onChange } = props;

  if (!steps) {
    return null;
  }

  const handleChange = (i: number) => (updatedStep: ApiStep, newIdx: number) => {
    if (i === newIdx) {
      // Just update the step in place
      const updatedSteps = steps.map((step, idx) =>
        idx === i ? updatedStep : step
      );
      onChange(updatedSteps);
    } else {
      // Move the step to a new position
      const updatedSteps = [...steps];
      const [removed] = updatedSteps.splice(i, 1);
      updatedSteps.splice(newIdx, 0, removed);
      onChange(updatedSteps);
    }
  };

  const handleDelete = (idx: number) => {
    const updatedSteps = steps.filter((_, i) => i !== idx);
    onChange(updatedSteps);
  };

  const addStep = () => {
    const cpy = [...steps];
    cpy.push(emptyStep());

    onChange(cpy);
  };

  if (steps.length === 0) {
    return (
      <Box sx={{ textAlign: "center", padding: 4 }}>
        <Typography variant="body1" color="text.secondary" sx={{ marginBottom: 2 }}>
          No steps defined yet
        </Typography>
        {editing && (
          <Button
            variant="contained"
            color="primary"
            startIcon={<AddIcon />}
            onClick={addStep}
          >
            Add First Step
          </Button>
        )}
      </Box>
    );
  }

  return (
    <Stack gap={2}>
      {steps.map((step, i) => (
        <Paper
          key={i}
          variant="outlined"
          sx={{
            padding: 2,
            position: "relative",
          }}
        >
          <Box sx={{ display: "flex", alignItems: "flex-start", gap: 2 }}>
            <Typography
              variant="h5"
              color="primary"
              sx={{ minWidth: 40, fontWeight: "bold" }}
            >
              {i + 1}
            </Typography>

            <Step
              flex={1}
              editing={editing}
              step={step}
              pos={{ idx: i, isLast: i === steps.length - 1 }}
              onChange={handleChange(i)}
              onDelete={() => handleDelete(i)}
            />
          </Box>
        </Paper>
      ))}

      {editing && (
        <Box sx={{ textAlign: "center", marginTop: 1 }}>
          <Button
            variant="outlined"
            color="primary"
            startIcon={<AddIcon />}
            onClick={addStep}
          >
            Add Step
          </Button>
        </Box>
      )}
    </Stack>
  );
};
