import { Stack, Typography } from "@mui/material";
import React, { useMemo } from "react";

interface Props {
  key?: React.Key;
  percentage?: number;
  handleChange?: (updatedPercentage: number) => void;
  showInfo?: boolean;
  size?: "xs" | "sm" | "lg";
}

const colorOn = "orange";
const colorOff = "lightblue";

export const Cycle: React.FC<Props> = (props) => {
  const { key, percentage, handleChange, showInfo = true, size = "lg" } = props;

  if (percentage === undefined) {
    return null;
  }

  const sqSize = useMemo(
    () => (size === "lg" ? "2em" : size === "sm" ? "1.5em" : "0.5em"),
    [size]
  );

  return (
    <Stack direction="row" gap={3} alignItems="center">
      <Stack direction="row">
        {Array.from(Array(10).keys()).map((_, i) => {
          const val = (i + 1) * 10;
          const exact = val === percentage;

          return (
            <Stack
              key={`${key}-${i}`}
              style={{
                backgroundColor: val <= percentage ? colorOn : colorOff,
                cursor: !handleChange ? "default" : "pointer",
              }}
              height={sqSize}
              width={sqSize}
              border={1}
              borderColor="gray"
              onClick={() => handleChange?.(exact ? i * 10 : val)}
            />
          );
        })}
      </Stack>

      {showInfo && <Typography>{percentage} %</Typography>}
    </Stack>
  );
};
