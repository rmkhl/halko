import { Stack, Typography } from "@mui/material";
import React, { useMemo } from "react";
import { Cycle } from "../../types/api";

interface Props {
  cycle: Cycle;
  handleChange?: (updatedStated: boolean[]) => void;
}

const colorOn = "orange";
const colorOff = "lightblue";

export const States: React.FC<Props> = (props) => {
  const { cycle, handleChange } = props;
  const { id, states } = cycle;

  const ratio = useMemo(
    () => (states.filter((s) => s).length / states.length) * 100,
    [states]
  );

  const handleClick = (idx: number) => (e: React.MouseEvent) => {
    const newStates = [...states];
    newStates[idx] = !newStates[idx];

    handleChange?.(newStates);
  };

  return (
    <Stack direction="row" gap={3} alignItems="center">
      <Stack direction="row">
        {states.map((s, i) => (
          <Stack
            key={`cycle-${id}-state-${i + 1}`}
            style={{
              backgroundColor: s ? colorOn : colorOff,
              cursor: !handleChange ? "default" : "pointer",
            }}
            height="2em"
            width="2em"
            border={1}
            borderColor="gray"
            onClick={handleClick(i)}
          />
        ))}
      </Stack>

      <Typography>{ratio} %</Typography>
    </Stack>
  );
};
