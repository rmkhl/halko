import { Stack, Typography } from "@mui/material";
import React, { useMemo } from "react";
import { Cycle } from "../../types/api";

interface Props {
  cycle: Cycle;
  handleChange?: (updatedStated: boolean[]) => void;
  showInfo?: boolean;
  size?: "sm" | "lg";
}

const colorOn = "orange";
const colorOff = "lightblue";

export const States: React.FC<Props> = (props) => {
  const { cycle, handleChange, showInfo = true, size = "lg" } = props;
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

  const sqSize = useMemo(() => (size === "lg" ? "2em" : "1.5em"), [size]);

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
            height={sqSize}
            width={sqSize}
            border={1}
            borderColor="gray"
            onClick={handleClick(i)}
          />
        ))}
      </Stack>

      {showInfo && <Typography>{ratio} %</Typography>}
    </Stack>
  );
};
