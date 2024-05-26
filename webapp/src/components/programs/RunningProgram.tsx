import React, { useMemo } from "react";
import { useGetRunningProgramQuery } from "../../store/services/executorApi";
import { Program } from "../../types/api";
import { Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

export const RunningProgram: React.FC = () => {
  const { t } = useTranslation();
  const { data } = useGetRunningProgramQuery(undefined, {
    pollingInterval: 30000,
  });

  const runningProgram = useMemo(
    () => (data ? (data as Program) : undefined),
    [data]
  );

  if (!runningProgram) {
    return <Typography>{t("programs.noRunning")}</Typography>;
  }

  return <Typography>{runningProgram.name}</Typography>;
};
