import React, { useMemo } from "react";
import { Phase } from "../../types/api";
import { Stack, Typography, styled } from "@mui/material";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Cycle } from "../cycles";
import { DeltaCycles } from "./DeltaCycles";

interface Props {
  phase: Phase;
}

export const PhaseRow: React.FC<Props> = (props) => {
  const { phase } = props;
  const navigate = useNavigate();

  const handleRowClick = () => navigate(`/phases/${phase.id}`);

  return (
    <PhaseRowStack direction="row" onClick={handleRowClick}>
      <Stack flex={1}>
        <Typography variant="h5">{phase.name}</Typography>
      </Stack>

      <Stack flex={2}>
        <CycleInfo phase={phase} />
      </Stack>
    </PhaseRowStack>
  );
};

const PhaseRowStack = styled(Stack)(() => ({
  cursor: "pointer",
  padding: "1em",
  borderRadius: "1em",
  alignItems: "start",
  "&:hover": {
    backgroundColor: "#666",
  },
}));

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
