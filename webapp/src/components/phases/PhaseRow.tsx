import React from "react";
import { Phase } from "../../types/api";
import { Stack, Typography, styled } from "@mui/material";
import { useNavigate } from "react-router-dom";
import { celsiusRange } from "../../util";
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
        <Typography>{t(`phases.cycles.${phase.cycleMode}`)}</Typography>
      </Stack>

      <Stack flex={1}>
        <Typography>
          {phase.validRange
            .map(
              (v) =>
                `${
                  t(`phases.validRange.${v.sensor}`) || v.sensor
                }: ${celsiusRange(v.above, v.below)}`
            )
            .join(", ")}
        </Typography>
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
