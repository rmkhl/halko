import React, { useMemo } from "react";
import { Phase } from "../../types/api";
import { Stack, Typography, styled } from "@mui/material";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Cycle } from "../cycles";
import { DeltaCycles } from "./DeltaCycles";
import { ClickableStack } from "../ClickableStack";

interface Props {
  phase: Phase;
  onSelectRow?: (phase: Phase) => void;
  selectable?: boolean;
}

export const PhaseRow: React.FC<Props> = (props) => {
  const { phase, onSelectRow, selectable = true } = props;
  const navigate = useNavigate();

  const handleRowClick = () => navigate(`/phases/${phase.name}`);

  return (
    <ClickableStack
      direction="row"
      onClick={
        !selectable
          ? undefined
          : onSelectRow
          ? () => onSelectRow(phase)
          : handleRowClick
      }
    >
      <Stack flex={1}>
        <Typography variant="h5">{phase.name}</Typography>
      </Stack>

      <Stack flex={2}>
        <CycleInfo phase={phase} />
      </Stack>
    </ClickableStack>
  );
};

interface CycleInfoProps {
  phase: Phase;
}

const CycleInfo: React.FC<CycleInfoProps> = (props) => {
  const { phase } = props;
  const { cycleMode, constantCycle, deltaCycles } = phase;

  const { t } = useTranslation();

  const cycleModeStr = useMemo(
    () => t(`phases.cycles.${cycleMode}`),
    [cycleMode, t]
  );

  return (
    <Stack direction="row">
      <Stack flex={1}>
        <Typography>{cycleModeStr}</Typography>
      </Stack>

      {cycleMode === "delta" && (
        <Stack flex={2}>
          <DeltaCycles deltaCycles={deltaCycles} size="sm" />
        </Stack>
      )}

      {cycleMode === "constant" && (
        <Stack flex={2} alignItems="center">
          <Cycle percentage={constantCycle} size="sm" />
        </Stack>
      )}
    </Stack>
  );
};
