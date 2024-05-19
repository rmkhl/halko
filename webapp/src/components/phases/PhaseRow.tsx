import React, { useMemo } from "react";
import { Phase } from "../../types/api";
import { Stack, Typography, styled } from "@mui/material";
import { useNavigate } from "react-router-dom";
import { celsius } from "../../util";
import { useTranslation } from "react-i18next";

interface Props {
  phase: Phase;
}

export const PhaseRow: React.FC<Props> = (props) => {
  const { phase } = props;
  const navigate = useNavigate();

  const { t } = useTranslation();

  const handleRowClick = () => navigate(`/phases/${phase.id}`);

  return (
    <PhaseRowStack direction="row" onClick={handleRowClick}>
      <Stack flex={1}>
        <Typography variant="h5">{phase.name}</Typography>
      </Stack>

      <Stack flex={1}>
        <CycleInfo phase={phase} />
      </Stack>
    </PhaseRowStack>
  );
};

const PhaseRowStack = styled(Stack)(() => ({
  cursor: "pointer",
  padding: "1em",
  borderRadius: "1em",
  alignItems: "center",
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

  switch (cycleMode) {
    case "delta":
      return (
        <Typography>
          {cycleModeStr}:{" "}
          {[...(deltaCycles || [])]
            .sort((a, b) => a.delta - b.delta)
            .map((d) => celsius(d.delta))
            .join(", ")}
        </Typography>
      );

    case "constant":
      return (
        <Typography>
          {cycleModeStr}: {celsius(constantCycle)}
        </Typography>
      );

    default:
      return null;
  }
};
