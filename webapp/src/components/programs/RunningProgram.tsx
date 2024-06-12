import React, { useEffect, useMemo } from "react";
import {
  useGetRunningProgramQuery,
  useStopRunningProgramMutation,
} from "../../store/services/executorApi";
import { Program } from "../../types/api";
import { Button, Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

interface PowerStatus {
  fan: number;
  heater: number;
  humidifier: number;
}

interface Temperatures {
  delta: number;
  material: number;
  oven: number;
}

interface Data {
  power_status: PowerStatus;
  temperatures: Temperatures;
}

interface RunningResponse {
  data: Data;
}

export const RunningProgram: React.FC = () => {
  const { t } = useTranslation();
  const { data } = useGetRunningProgramQuery(undefined, {
    pollingInterval: 30000,
    skipPollingIfUnfocused: true,
  });

  const [stopProgram] = useStopRunningProgramMutation();

  const runningProgram = useMemo(
    () => (data ? (data as RunningResponse) : undefined),
    [data]
  );

  if (!runningProgram) {
    return <Typography>{t("programs.noRunning")}</Typography>;
  }

  return (
    <Stack direction="row">
      <Typography>{runningProgram.data.temperatures.material}</Typography>

      <Button onClick={() => stopProgram("")}>Stop</Button>
    </Stack>
  );
};
