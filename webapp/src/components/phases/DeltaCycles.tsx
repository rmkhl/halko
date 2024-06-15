import React from "react";
import { DeltaCycle as ApiDeltaCycle } from "../../types/api";
import { DeltaCycle } from "./DeltaCycle";
import { Stack } from "@mui/material";

interface Props {
  deltaCycles?: ApiDeltaCycle[];
  size?: "sm" | "lg";
  onChange?: (cycles: ApiDeltaCycle[]) => void;
}

export const DeltaCycles: React.FC<Props> = (props) => {
  const { deltaCycles, onChange, size = "lg" } = props;

  const handleChange =
    (idx: number, delta: "above" | "below") => (percentage: number) => {
      onChange?.(
        deltaCycles?.map((c, i) =>
          i === idx
            ? {
                delta: c.delta,
                above: delta === "above" ? percentage : c.above,
                below: delta === "below" ? percentage : c.below,
              }
            : { ...c }
        ) || []
      );
    };

  return (
    !!deltaCycles && (
      <Stack gap={size === "lg" ? 2 : undefined}>
        {deltaCycles?.map((curr, i, cycles) => {
          const prev = cycles[i - 1];
          const next = cycles[i + 1];

          return (
            <DeltaCycle
              key={`deltaCycle-${curr.delta}`}
              prev={prev}
              curr={curr}
              next={next}
              onChangeAbove={handleChange(i + 1, "above")}
              onChangeBelow={!!next ? handleChange(i, "below") : undefined}
              size={size}
            />
          );
        })}
      </Stack>
    )
  );
};
