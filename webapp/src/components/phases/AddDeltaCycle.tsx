import React, { useEffect, useMemo, useState } from "react";
import {
  Cycle as ApiCycle,
  DeltaCycle as ApiDeltaCycle,
} from "../../types/api";
import { Button, Slider, Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { celsius } from "../../util";
import ArrowDownwardRoundedIcon from "@mui/icons-material/ArrowDownwardRounded";
import ArrowUpwardRoundedIcon from "@mui/icons-material/ArrowUpwardRounded";
import { Cycle } from "../cycles/Cycle";
import { CycleSelector } from "./CycleSelector";

interface Props {
  existingDeltaCycles?: ApiDeltaCycle[];
  onSelect: (deltaCycle: ApiDeltaCycle) => void;
}

export const AddDeltaCycle: React.FC<Props> = (props) => {
  const { existingDeltaCycles, onSelect } = props;
  const [existingDeltas, setExistingDeltas] = useState<Set<number>>(new Set());
  const [delta, setDelta] = useState(50);
  const [below, setBelow] = useState<ApiCycle | undefined>();
  const [above, setAbove] = useState<ApiCycle | undefined>();

  const { t } = useTranslation();

  useEffect(() => {
    const deltas = new Set<number>();

    existingDeltaCycles?.forEach((d) => deltas.add(d.delta));

    setExistingDeltas(deltas);
  }, [existingDeltaCycles]);

  const handleDeltaChange = (_: Event, value: number | number[]) => {
    if (Array.isArray(value)) {
      return;
    }

    setDelta(value);
  };

  const select =
    (apiCycleSetter: (apiCycle: ApiCycle | undefined) => void) =>
    (apiCycle: ApiCycle) => {
      apiCycleSetter(apiCycle);
    };

  const addApiCycle = () => {
    if (!below || !above || existingDeltas.has(delta)) {
      return;
    }

    const newDeltaCycle: ApiDeltaCycle = {
      delta,
      above,
      below,
    };

    onSelect(newDeltaCycle);
  };

  const selectionIsValid = useMemo(
    () => !!below && !!above && !existingDeltas.has(delta),
    [below, above, existingDeltas, delta]
  );

  return (
    <Stack>
      <Stack>
        <Typography>
          {t("phases.cycles.delta")}: {celsius(delta)}
        </Typography>

        <Slider
          value={delta}
          step={1}
          getAriaValueText={celsius}
          marks={rangeMarks}
          max={200}
          min={0}
          track={false}
          valueLabelDisplay="auto"
          onChange={handleDeltaChange}
          disableSwap
        />
      </Stack>

      <Stack>
        <Stack direction="row">
          <Stack flex={1} alignItems="center">
            {below === undefined ? (
              <CycleSelector onSelect={select(setBelow)} />
            ) : (
              <Cycle cycle={below} showInfo={false} size="sm" />
            )}
          </Stack>

          <Stack flex={0.5} />

          <Stack flex={1} />
        </Stack>

        <Stack direction="row">
          <Stack flex={1} alignItems="center">
            <ArrowUpwardRoundedIcon />
          </Stack>

          <Stack flex={0.5} alignItems="center">
            <Typography>{celsius(delta)}</Typography>
          </Stack>

          <Stack flex={1} alignItems="center">
            <ArrowDownwardRoundedIcon />
          </Stack>
        </Stack>

        <Stack direction="row">
          <Stack flex={1} />

          <Stack flex={0.5} />

          <Stack flex={1} alignItems="center">
            {above === undefined ? (
              <CycleSelector onSelect={select(setAbove)} />
            ) : (
              <Cycle cycle={above} showInfo={false} size="sm" />
            )}
          </Stack>
        </Stack>
      </Stack>

      <Button
        onClick={addApiCycle}
        disabled={!selectionIsValid}
        color="success"
      >
        {t("phases.cycles.addDeltaCycle")}
      </Button>
    </Stack>
  );
};

const rangeMarks = [
  {
    value: 0,
    label: "0°C",
  },
  {
    value: 25,
    label: "25°C",
  },
  {
    value: 50,
    label: "50°C",
  },
  {
    value: 75,
    label: "75°C",
  },
  {
    value: 100,
    label: "100°C",
  },
  {
    value: 125,
    label: "125°C",
  },
  {
    value: 150,
    label: "150°C",
  },
  {
    value: 175,
    label: "175°C",
  },
  {
    value: 200,
    label: "200°C",
  },
];
