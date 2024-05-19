import React, { useMemo } from "react";
import { DeltaCycle as ApiDeltaCycle } from "../../types/api";
import { DeltaCycle } from "./DeltaCycle";
import { Stack } from "@mui/material";
import { useTranslation } from "react-i18next";

interface Props {
  editing?: boolean;
  deltaCycles?: ApiDeltaCycle[];
  onChange: (cycles: ApiDeltaCycle[]) => void;
}

export const DeltaCycles: React.FC<Props> = (props) => {
  const { editing, deltaCycles, onChange } = props;
  const { t } = useTranslation();
  const addDeltaCycleStr = useMemo(() => t("phases.cycles.addDeltaCycle"), [t]);

  const handleChange =
    (idx: number, delta: "above" | "below") => (percentage: number) => {
      console.log(idx);
      onChange(
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
    !!deltaCycles &&
    deltaCycles.length === 13 && (
      <Stack gap={2}>
        {deltaCycles?.map((curr, i, cycles) => {
          const prev = cycles[i - 1];
          const next = cycles[i + 1];
          console.log(
            "prev",
            i - 1,
            prev,
            "curr",
            i,
            curr,
            "next",
            i + 1,
            next
          );

          return (
            <DeltaCycle
              key={`deltaCycle-${curr.delta}`}
              prev={prev}
              curr={curr}
              next={next}
              onChangeAbove={handleChange(i + 1, "above")}
              onChangeBelow={!!next ? handleChange(i, "below") : undefined}
            />
          );
        })}
      </Stack>
    )
  );
};
